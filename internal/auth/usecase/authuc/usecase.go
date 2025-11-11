package authuc

import (
	"chatx-01/internal/auth/domain"
	"chatx-01/pkg/errjon"
	"chatx-01/pkg/tokenx"
	"context"
)

type useCase struct {
	userRepo       domain.UserRepository
	passwordHasher domain.PasswordHasher
	tokenGenerator tokenx.Generator
}

// New creates a new auth use case.
func New(
	userRepo domain.UserRepository,
	passwordHasher domain.PasswordHasher,
	tokenGenerator tokenx.Generator,
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
		return nil, errjon.ReplaceOn(err, errjon.ErrNotFound, domain.ErrInvalidCredentials)
	}

	// Compare password
	if err := uc.passwordHasher.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, errjon.Wrap(op, domain.ErrInvalidCredentials)
	}

	// Generate tokens
	accessToken, err := uc.tokenGenerator.Generate(user.ID, string(user.Role), tokenx.TokenTypeAccess)
	if err != nil {
		return nil, errjon.Wrap(op, err)
	}

	refreshToken, err := uc.tokenGenerator.Generate(user.ID, string(user.Role), tokenx.TokenTypeRefresh)
	if err != nil {
		return nil, errjon.Wrap(op, err)
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
