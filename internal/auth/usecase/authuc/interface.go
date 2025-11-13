package authuc

import (
	"chatx-01-backend/pkg/errs"
	"chatx-01-backend/pkg/val"
	"context"
)

type UseCase interface {
	Login(ctx context.Context, req LoginReq) (*LoginResp, error)
	Logout(ctx context.Context, req LogoutReq) error
}

// Login request/response.
type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req LoginReq) Validate() error {
	var verr error

	if err := val.ValidateEmail(req.Email); err != nil {
		verr = errs.AddFieldError(verr, "email", err.Error())
	}
	if req.Password == "" {
		verr = errs.AddFieldError(verr, "password", "password is required")
	}

	return verr
}

type LoginResp struct {
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	ImagePath    string `json:"image_path,omitempty"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Logout request.
type LogoutReq struct {
}

func (req LogoutReq) Validate() error {
	return nil
}
