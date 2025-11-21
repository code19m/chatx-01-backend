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
	tokenService   *token.Service
}

func New(
	userRepo domain.UserRepository,
	passwordHasher hasher.Hasher,
	tokenService *token.Service,
) UseCase {
	return &useCase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		tokenService:   tokenService,
	}
}

func (uc *useCase) Login(ctx context.Context, req LoginReq) (*LoginResp, error) {
	const op = "authuc.Login"

	user, err := uc.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errs.ReplaceOn(err, errs.ErrNotFound, domain.ErrInvalidCredentials)
	}

	err = uc.passwordHasher.Compare(user.PasswordHash, req.Password)
	if err != nil {
		return nil, errs.Wrap(op, domain.ErrInvalidCredentials)
	}

	// Generate and store access token in Redis
	accessToken, err := uc.tokenService.GenerateAndStore(ctx, user.ID, user.Role.String(), token.TokenTypeAccess)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	// Generate and store refresh token in Redis
	refreshToken, err := uc.tokenService.GenerateAndStore(ctx, user.ID, user.Role.String(), token.TokenTypeRefresh)
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
	const op = "authuc.Logout"

	// Revoke access token
	if req.AccessToken != "" {
		err := uc.tokenService.Revoke(ctx, req.AccessToken)
		if err != nil {
			return errs.Wrap(op, err)
		}
	}

	return nil
}
