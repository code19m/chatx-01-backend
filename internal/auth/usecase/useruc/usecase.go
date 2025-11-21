package useruc

import (
	"bytes"
	"chatx-01-backend/internal/auth/domain"
	"chatx-01-backend/internal/events"
	"chatx-01-backend/internal/portal/auth"
	"chatx-01-backend/pkg/errs"
	"chatx-01-backend/pkg/filestore"
	"chatx-01-backend/pkg/hasher"
	"chatx-01-backend/pkg/kafka"
	"chatx-01-backend/pkg/token"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type useCase struct {
	userRepo       domain.UserRepository
	passwordHasher hasher.Hasher
	fileStore      filestore.Store
	authPr         auth.Portal
	eventProducer  *kafka.Producer
	tokenService   *token.Service
}

func New(
	userRepo domain.UserRepository,
	passwordHasher hasher.Hasher,
	fileStore filestore.Store,
	authPr auth.Portal,
	eventProducer *kafka.Producer,
	tokenService *token.Service,
) UseCase {
	return &useCase{
		userRepo,
		passwordHasher,
		fileStore,
		authPr,
		eventProducer,
		tokenService,
	}
}

func (uc *useCase) CreateUser(ctx context.Context, req CreateUserReq) (*CreateUserResp, error) {
	const op = "useruc.CreateUser"

	// Send user registration event to Kafka with plain password before hashing
	event := events.UserRegisteredEvent{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	eventData, err := event.Marshal()
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	err = uc.eventProducer.SendMessage(ctx, &kafka.Message{
		Key:   []byte(req.Email),
		Value: eventData,
	})
	if err != nil {
		return nil, errs.Wrap(op, fmt.Errorf("failed to send registration event: %w", err))
	}

	// Now hash the password and create the user
	passwordHash, err := uc.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	user := &domain.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Role:         domain.RoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
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

func (uc *useCase) CreateSuperUser(ctx context.Context, req CreateSuperUserReq) (*CreateSuperUserResp, error) {
	const op = "useruc.CreateSuperUser"

	passwordHash, err := uc.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	user := &domain.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Role:         domain.RoleAdmin,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, errs.ReplaceOn(
			err,
			errs.ErrAlreadyExists,
			errs.NewConflictError("email", "email already exists"),
		)
	}

	return &CreateSuperUserResp{
		UserID: user.ID,
	}, nil
}

func (uc *useCase) DeleteUser(ctx context.Context, req DeleteUserReq) error {
	const op = "useruc.DeleteUser"

	// Get user to ensure they exist
	_, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("user_id", "user not found"))
	}

	// Revoke all user tokens BEFORE deleting
	err = uc.tokenService.RevokeAllUserTokens(ctx, req.UserID)
	if err != nil {
		slog.Error("failed to revoke user tokens", "user_id", req.UserID, "error", err)
		// Don't fail deletion - continue
	}

	// Delete user
	err = uc.userRepo.Delete(ctx, req.UserID)
	if err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (uc *useCase) GetUser(ctx context.Context, req GetUserReq) (*GetUserResp, error) {
	const op = "useruc.GetUser"

	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("user_id", "user not found"))
	}

	return &GetUserResp{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		ImagePath: user.ImagePath,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (uc *useCase) GetUsersList(ctx context.Context, req GetUsersListReq) (*GetUsersListResp, error) {
	const op = "useruc.GetUsersList"

	offset := req.Page * req.Limit
	var users []*domain.User
	var total int
	var err error

	if req.Search != "" {
		users, total, err = uc.userRepo.SearchByUsernameWithCount(ctx, req.Search, offset, req.Limit)
	} else {
		users, total, err = uc.userRepo.ListWithCount(ctx, offset, req.Limit)
	}

	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	userItems := make([]UserListItem, len(users))
	for i, user := range users {
		userItems[i] = UserListItem{
			UserID:    user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
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

	au, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	user, err := uc.userRepo.GetByID(ctx, au.ID)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &GetMeResp{
		UserID:    user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		ImagePath: user.ImagePath,
	}, nil
}

func (uc *useCase) ChangePassword(ctx context.Context, req ChangePasswordReq) error {
	const op = "useruc.ChangePassword"

	au, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return errs.Wrap(op, err)
	}

	user, err := uc.userRepo.GetByID(ctx, au.ID)
	if err != nil {
		return errs.Wrap(op, err)
	}

	err = uc.passwordHasher.Compare(user.PasswordHash, req.OldPassword)
	if err != nil {
		return errs.Wrap(op, errs.NewNotFoundError("old_password", domain.ErrIncorrectPassword.Error()))
	}

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

	au, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	exists, err := uc.fileStore.Exists(ctx, req.ImagePath)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	if !exists {
		return nil, errs.Wrap(op, errs.NewNotFoundError("image_path", "file does not exist"))
	}

	contentType, err := uc.fileStore.GetContentType(ctx, req.ImagePath)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	if !slices.Contains([]string{
		"image/jpeg",
		"image/jpg",
		"image/png",
	}, strings.ToLower(contentType)) {
		return nil, errs.Wrap(op, errs.NewValidationError("file must be a JPEG or PNG image"))
	}

	user, err := uc.userRepo.GetByID(ctx, au.ID)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	user.ImagePath = &req.ImagePath
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &ChangeImageResp{
		ImagePath: user.ImagePath,
	}, nil
}

func (uc *useCase) UploadImage(ctx context.Context, req UploadImageReq) (*UploadImageResp, error) {
	const op = "useruc.UploadImage"

	au, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	if !slices.Contains([]string{
		"image/jpeg",
		"image/jpg",
		"image/png",
	}, strings.ToLower(req.ContentType)) {
		return nil, errs.Wrap(op, errs.NewValidationError("file must be a JPEG or PNG image"))
	}

	ext := filepath.Ext(req.FileName)
	imagePath := fmt.Sprintf("users/%d/profile%s", au.ID, ext)

	reader := bytes.NewReader(req.File)
	err = uc.fileStore.Upload(ctx, imagePath, reader, req.Size, req.ContentType)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &UploadImageResp{
		ImagePath: imagePath,
	}, nil
}

func (uc *useCase) DownloadImage(ctx context.Context, req DownloadImageReq) (*DownloadImageResp, error) {
	const op = "useruc.DownloadImage"

	exists, err := uc.fileStore.Exists(ctx, req.ImagePath)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	if !exists {
		return nil, errs.Wrap(op, errs.NewNotFoundError("image_path", "file does not exist"))
	}

	contentType, err := uc.fileStore.GetContentType(ctx, req.ImagePath)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	reader, err := uc.fileStore.Download(ctx, req.ImagePath)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	defer reader.Close()

	fileData, err := io.ReadAll(reader)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	fileName := filepath.Base(req.ImagePath)

	return &DownloadImageResp{
		File:        fileData,
		ContentType: contentType,
		FileName:    fileName,
	}, nil
}
