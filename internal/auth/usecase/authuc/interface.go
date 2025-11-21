package authuc

import (
	"chatx-01-backend/internal/auth/domain"
	"chatx-01-backend/pkg/errs"
	"chatx-01-backend/pkg/val"
	"context"
)

type UseCase interface {
	Login(ctx context.Context, req LoginReq) (*LoginResp, error)
	Logout(ctx context.Context, req LogoutReq) error
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (req LoginReq) Validate() error {
	var verr error

	if err := val.ValidateUsername(req.Username); err != nil {
		verr = errs.AddFieldError(verr, "username", err.Error())
	}
	if req.Password == "" {
		verr = errs.AddFieldError(verr, "password", "password is required")
	}

	return verr
}

type LoginResp struct {
	UserID       int             `json:"user_id"`
	Username     string          `json:"username"`
	Email        string          `json:"email"`
	Role         domain.UserRole `json:"role"`
	ImagePath    *string         `json:"image_path"`
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
}

type LogoutReq struct {
	AccessToken string `json:"-"` // Not from JSON, set by handler
}

func (req LogoutReq) Validate() error {
	return nil
}
