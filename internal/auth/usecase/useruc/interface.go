package useruc

import (
	"chatx-01-backend/internal/auth/domain"
	"chatx-01-backend/pkg/errs"
	"chatx-01-backend/pkg/val"
	"context"
)

type UseCase interface {
	CreateUser(ctx context.Context, req CreateUserReq) (*CreateUserResp, error)
	CreateSuperUser(ctx context.Context, req CreateSuperUserReq) (*CreateSuperUserResp, error)
	DeleteUser(ctx context.Context, req DeleteUserReq) error
	GetUser(ctx context.Context, req GetUserReq) (*GetUserResp, error)
	GetUsersList(ctx context.Context, req GetUsersListReq) (*GetUsersListResp, error)
	GetMe(ctx context.Context, req GetMeReq) (*GetMeResp, error)
	ChangePassword(ctx context.Context, req ChangePasswordReq) error
	ChangeImage(ctx context.Context, req ChangeImageReq) (*ChangeImageResp, error)
	UploadImage(ctx context.Context, req UploadImageReq) (*UploadImageResp, error)
	DownloadImage(ctx context.Context, req DownloadImageReq) (*DownloadImageResp, error)
}

type CreateUserReq struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (req CreateUserReq) Validate() error {
	var verr error

	if err := val.ValidateEmail(req.Email); err != nil {
		verr = errs.AddFieldError(verr, "email", err.Error())
	}
	if err := val.ValidateUsername(req.Username); err != nil {
		verr = errs.AddFieldError(verr, "username", err.Error())
	}

	return verr
}

type CreateUserResp struct {
	UserID int `json:"user_id"`
}

type CreateSuperUserReq struct {
	Email    string
	Username string
	Password string
}

func (req CreateSuperUserReq) Validate() error {
	var verr error

	if err := val.ValidateEmail(req.Email); err != nil {
		verr = errs.AddFieldError(verr, "email", err.Error())
	}
	if err := val.ValidateUsername(req.Username); err != nil {
		verr = errs.AddFieldError(verr, "username", err.Error())
	}
	if req.Password == "" {
		verr = errs.AddFieldError(verr, "password", "password is required")
	}
	if len(req.Password) < 8 {
		verr = errs.AddFieldError(verr, "password", "password must be at least 8 characters")
	}

	return verr
}

type CreateSuperUserResp struct {
	UserID int
}

type DeleteUserReq struct {
	UserID int `path:"user_id"`
}

func (req DeleteUserReq) Validate() error {
	var verr error

	if req.UserID <= 0 {
		verr = errs.AddFieldError(verr, "user_id", "invalid user id")
	}

	return verr
}

type GetUserReq struct {
	UserID int `path:"user_id"`
}

func (req GetUserReq) Validate() error {
	var verr error

	if req.UserID <= 0 {
		verr = errs.AddFieldError(verr, "user_id", "invalid user id")
	}

	return verr
}

type GetUserResp struct {
	UserID    int             `json:"user_id"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	Role      domain.UserRole `json:"role"`
	ImagePath *string         `json:"image_path"`
	CreatedAt string          `json:"created_at"`
}

type GetUsersListReq struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Search string `query:"search"`
}

func (req GetUsersListReq) Validate() error {
	var verr error

	if req.Page < 0 {
		verr = errs.AddFieldError(verr, "page", "page must be non-negative")
	}
	if req.Limit <= 0 || req.Limit > 100 {
		verr = errs.AddFieldError(verr, "limit", "limit must be between 1 and 100")
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
	UserID    int             `json:"user_id"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	Role      domain.UserRole `json:"role"`
	ImagePath *string         `json:"image_path"`
	CreatedAt string          `json:"created_at"`
}

type GetMeReq struct{}

func (req GetMeReq) Validate() error {
	return nil
}

type GetMeResp struct {
	UserID    int             `json:"user_id"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	Role      domain.UserRole `json:"role"`
	ImagePath *string         `json:"image_path"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (req ChangePasswordReq) Validate() error {
	var verr error

	if req.OldPassword == "" {
		verr = errs.AddFieldError(verr, "old_password", "old password is required")
	}
	if req.NewPassword == "" {
		verr = errs.AddFieldError(verr, "new_password", "new password is required")
	}
	if len(req.NewPassword) < 8 {
		verr = errs.AddFieldError(verr, "new_password", "password must be at least 8 characters")
	}

	return verr
}

type ChangeImageReq struct {
	ImagePath string `json:"image_path"`
}

func (req ChangeImageReq) Validate() error {
	var verr error

	if req.ImagePath == "" {
		verr = errs.AddFieldError(verr, "image_path", "image path is required")
	}

	return verr
}

type ChangeImageResp struct {
	ImagePath *string `json:"image_path"`
}

type UploadImageReq struct {
	File        []byte `json:"-"`
	FileName    string `json:"-"`
	ContentType string `json:"-"`
	Size        int64  `json:"-"`
}

func (req UploadImageReq) Validate() error {
	var verr error

	if len(req.File) == 0 {
		verr = errs.AddFieldError(verr, "file", "file is required")
	}
	if req.FileName == "" {
		verr = errs.AddFieldError(verr, "file_name", "file name is required")
	}
	if req.Size <= 0 {
		verr = errs.AddFieldError(verr, "size", "invalid file size")
	}

	return verr
}

type UploadImageResp struct {
	ImagePath string `json:"image_path"`
}

type DownloadImageReq struct {
	ImagePath string `path:"image_path"`
}

func (req DownloadImageReq) Validate() error {
	var verr error

	if req.ImagePath == "" {
		verr = errs.AddFieldError(verr, "image_path", "image path is required")
	}

	return verr
}

type DownloadImageResp struct {
	File        []byte
	ContentType string
	FileName    string
}
