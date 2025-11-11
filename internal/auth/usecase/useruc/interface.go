package useruc

import (
	"chatx-01/pkg/errjon"
	"chatx-01/pkg/val"
	"context"
)

type UseCase interface {
	CreateUser(ctx context.Context, req CreateUserReq) (*CreateUserResp, error)
	DeleteUser(ctx context.Context, req DeleteUserReq) error
	GetUser(ctx context.Context, req GetUserReq) (*GetUserResp, error)
	GetUsersList(ctx context.Context, req GetUsersListReq) (*GetUsersListResp, error)
	GetMe(ctx context.Context, req GetMeReq) (*GetMeResp, error)
	ChangePassword(ctx context.Context, req ChangePasswordReq) error
	ChangeImage(ctx context.Context, req ChangeImageReq) (*ChangeImageResp, error)
}

type CreateUserReq struct {
	Email    string
	Username string
	Password string
}

func (req CreateUserReq) Validate() error {
	var verr error

	if err := val.ValidateEmail(req.Email); err != nil {
		verr = errjon.AddFieldError(verr, "email", err.Error())
	}
	if err := val.ValidateUsername(req.Username); err != nil {
		verr = errjon.AddFieldError(verr, "username", err.Error())
	}

	return verr
}

type CreateUserResp struct {
	UserID int `json:"user_id"`
}

// DeleteUser request.
type DeleteUserReq struct {
	UserID int `path:"userId"`
}

func (req DeleteUserReq) Validate() error {
	var verr error

	if req.UserID <= 0 {
		verr = errjon.AddFieldError(verr, "user_id", "invalid user id")
	}

	return verr
}

// GetUser request/response.
type GetUserReq struct {
	UserID int `path:"userId"`
}

func (req GetUserReq) Validate() error {
	var verr error

	if req.UserID <= 0 {
		verr = errjon.AddFieldError(verr, "user_id", "invalid user id")
	}

	return verr
}

type GetUserResp struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ImagePath string `json:"image_path,omitempty"`
	CreatedAt string `json:"created_at"`
}

// GetUsersList request/response.
type GetUsersListReq struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

func (req GetUsersListReq) Validate() error {
	var verr error

	if req.Page < 0 {
		verr = errjon.AddFieldError(verr, "page", "page must be non-negative")
	}
	if req.Limit <= 0 || req.Limit > 100 {
		verr = errjon.AddFieldError(verr, "limit", "limit must be between 1 and 100")
	}

	return verr
}

type GetUsersListResp struct {
	Users []UserListItem `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

type UserListItem struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ImagePath string `json:"image_path,omitempty"`
	CreatedAt string `json:"created_at"`
}

// GetMe request/response.
type GetMeReq struct {
}

func (req GetMeReq) Validate() error {
	return nil
}

type GetMeResp struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ImagePath string `json:"image_path,omitempty"`
}

// ChangePassword request.
type ChangePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (req ChangePasswordReq) Validate() error {
	var verr error

	if req.OldPassword == "" {
		verr = errjon.AddFieldError(verr, "old_password", "old password is required")
	}
	if req.NewPassword == "" {
		verr = errjon.AddFieldError(verr, "new_password", "new password is required")
	}
	if len(req.NewPassword) < 8 {
		verr = errjon.AddFieldError(verr, "new_password", "password must be at least 8 characters")
	}

	return verr
}

// ChangeImage request/response.
type ChangeImageReq struct {
	ImagePath string `json:"image_path"`
}

func (req ChangeImageReq) Validate() error {
	var verr error

	if req.ImagePath == "" {
		verr = errjon.AddFieldError(verr, "image_path", "image path is required")
	}

	return verr
}

type ChangeImageResp struct {
	ImagePath string `json:"image_path"`
}
