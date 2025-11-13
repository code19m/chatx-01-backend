package domain

import (
	"context"
	"time"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

func (r UserRole) IsValid() bool {
	return r == RoleAdmin || r == RoleUser
}

type User struct {
	ID           int
	Email        string
	Username     string
	PasswordHash string
	Role         UserRole
	ImagePath    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int) error
	ListWithCount(ctx context.Context, offset, limit int) ([]*User, int, error)
}

// PasswordHasher defines the interface for password hashing operations.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

// FileStore defines the interface for file storage operations.
type FileStore interface {
	Exists(ctx context.Context, path string) (bool, error)
	GetContentType(ctx context.Context, path string) (string, error)
}
