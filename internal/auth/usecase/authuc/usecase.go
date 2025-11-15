package authuc

import (
	"chatx-01-backend/internal/auth/domain"
	"chatx-01-backend/pkg/errs"
	"chatx-01-backend/pkg/hasher"
	"chatx-01-backend/pkg/token"
	"context"
)

type useCase struct {
	userRepo       domain.UserRepository
	passwordHasher hasher.Hasher
	tokenGenerator token.Generator
}

func New(
	userRepo domain.UserRepository,
	passwordHasher hasher.Hasher,
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

	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errs.ReplaceOn(err, errs.ErrNotFound, domain.ErrInvalidCredentials)
	}

	err = uc.passwordHasher.Compare(user.PasswordHash, req.Password)
	if err != nil {
		return nil, errs.Wrap(op, domain.ErrInvalidCredentials)
	}

	accessToken, err := uc.tokenGenerator.Generate(user.ID, user.Role.String(), token.TokenTypeAccess)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	refreshToken, err := uc.tokenGenerator.Generate(user.ID, user.Role.String(), token.TokenTypeRefresh)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &LoginResp{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Role:         user.Role,
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
