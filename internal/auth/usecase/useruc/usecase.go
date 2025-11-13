package useruc

import (
	"chatx-01-backend/internal/auth/domain"
	"chatx-01-backend/internal/portal/auth"
	"chatx-01-backend/pkg/errs"
	"context"
	"strings"
	"time"
)

type useCase struct {
	userRepo       domain.UserRepository
	passwordHasher domain.PasswordHasher
	fileStore      domain.FileStore
	authPr         auth.Auth
}

// New creates a new user use case.
func New(
	userRepo domain.UserRepository,
	passwordHasher domain.PasswordHasher,
	fileStore domain.FileStore,
	authPr auth.Auth,
) UseCase {
	return &useCase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		fileStore:      fileStore,
		authPr:         authPr,
	}
}

func (uc *useCase) CreateUser(ctx context.Context, req CreateUserReq) (*CreateUserResp, error) {
	const op = "useruc.CreateUser"

	// Check if user with email already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errs.Wrap(op, errs.NewConflictError("email", "email already exists"))
	}

	// Hash password
	passwordHash, err := uc.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	// Create user
	user := &domain.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Role:         domain.RoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		// Check if repository returned AlreadyExists error (race condition)
		return nil, errs.ReplaceOn(
			err,
			errs.ErrAlreadyExists,
			errs.NewConflictError("email", "email already exists"),
		)
	}

	return &CreateUserResp{
		UserID: user.ID,
	}, nil
}

func (uc *useCase) DeleteUser(ctx context.Context, req DeleteUserReq) error {
	const op = "useruc.DeleteUser"

	// Check if user exists - this is from user input, so replace with NotFoundError
	_, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("user_id", "user not found"))
	}

	if err := uc.userRepo.Delete(ctx, req.UserID); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (uc *useCase) GetUser(ctx context.Context, req GetUserReq) (*GetUserResp, error) {
	const op = "useruc.GetUser"

	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		// This is from user input, so replace with NotFoundError
		return nil, errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("user_id", "user not found"))
	}

	return &GetUserResp{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      string(user.Role),
		ImagePath: user.ImagePath,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (uc *useCase) GetUsersList(ctx context.Context, req GetUsersListReq) (*GetUsersListResp, error) {
	const op = "useruc.GetUsersList"

	offset := req.Page * req.Limit
	users, total, err := uc.userRepo.ListWithCount(ctx, offset, req.Limit)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	userItems := make([]UserListItem, len(users))
	for i, user := range users {
		userItems[i] = UserListItem{
			UserID:    user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      string(user.Role),
			ImagePath: user.ImagePath,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		}
	}

	return &GetUsersListResp{
		Users: userItems,
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

func (uc *useCase) GetMe(ctx context.Context, req GetMeReq) (*GetMeResp, error) {
	const op = "useruc.GetMe"

	authUser, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	user, err := uc.userRepo.GetByID(ctx, authUser.ID)
	if err != nil {
		// This is from our internal logic (JWT token), not user input, so don't replace
		return nil, errs.Wrap(op, err)
	}

	return &GetMeResp{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      string(user.Role),
		ImagePath: user.ImagePath,
	}, nil
}

func (uc *useCase) ChangePassword(ctx context.Context, req ChangePasswordReq) error {
	const op = "useruc.ChangePassword"

	authUser, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return errs.Wrap(op, err)
	}

	user, err := uc.userRepo.GetByID(ctx, authUser.ID)
	if err != nil {
		// This is from our internal logic (JWT token), not user input, so don't replace
		return errs.Wrap(op, err)
	}

	// Verify old password
	if err := uc.passwordHasher.Compare(user.PasswordHash, req.OldPassword); err != nil {
		return errs.Wrap(op, errs.NewNotFoundError("old_password", domain.ErrIncorrectPassword.Error()))
	}

	// Hash new password
	newPasswordHash, err := uc.passwordHasher.Hash(req.NewPassword)
	if err != nil {
		return errs.Wrap(op, err)
	}

	user.PasswordHash = newPasswordHash
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (uc *useCase) ChangeImage(ctx context.Context, req ChangeImageReq) (*ChangeImageResp, error) {
	const op = "useruc.ChangeImage"

	authUser, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	// Check if file exists
	exists, err := uc.fileStore.Exists(ctx, req.ImagePath)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	if !exists {
		return nil, errs.Wrap(op, errs.NewNotFoundError("image_path", "file does not exist"))
	}

	// Check content type
	contentType, err := uc.fileStore.GetContentType(ctx, req.ImagePath)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	contentType = strings.ToLower(contentType)
	if contentType != "image/jpeg" && contentType != "image/jpg" && contentType != "image/png" {
		return nil, errs.Wrap(op, errs.NewValidationError("file must be a JPEG or PNG image"))
	}

	// Update user image
	user, err := uc.userRepo.GetByID(ctx, authUser.ID)
	if err != nil {
		// This is from our internal logic (JWT token), not user input, so don't replace
		return nil, errs.Wrap(op, err)
	}

	user.ImagePath = req.ImagePath
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &ChangeImageResp{
		ImagePath: user.ImagePath,
	}, nil
}
