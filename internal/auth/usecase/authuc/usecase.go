package authuc

import (
	"chatx-01/internal/auth/domain"
	"chatx-01/pkg/errs"
	"chatx-01/pkg/token"
	"context"
)

type useCase struct {
	userRepo       domain.UserRepository
	passwordHasher domain.PasswordHasher
	tokenGenerator token.Generator
}

// New creates a new auth use case.
func New(
	userRepo domain.UserRepository,
	passwordHasher domain.PasswordHasher,
	tokenGenerator token.Generator,
) UseCase {
	return &useCase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenGenerator: tokenGenerator,
	}
}

func (uc *useCase) Login(ctx context.Context, req LoginReq) (*LoginResp, error) {
	const op = "authuc.Login"

	// Get user by email - if not found due to user input, replace with domain error
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errs.ReplaceOn(err, errs.ErrNotFound, domain.ErrInvalidCredentials)
	}

	// Compare password
	if err := uc.passwordHasher.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, errs.Wrap(op, domain.ErrInvalidCredentials)
	}

	// Generate tokens
	accessToken, err := uc.tokenGenerator.Generate(user.ID, string(user.Role), token.TokenTypeAccess)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	refreshToken, err := uc.tokenGenerator.Generate(user.ID, string(user.Role), token.TokenTypeRefresh)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &LoginResp{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Role:         string(user.Role),
		ImagePath:    user.ImagePath,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (uc *useCase) Logout(ctx context.Context, req LogoutReq) error {
	// For now, we don't do server-side session invalidation
	// Just return success
	return nil
}
