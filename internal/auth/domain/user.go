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

func (r UserRole) String() string {
	return string(r)
}

type User struct {
	ID           int
	Email        string
	Username     string
	PasswordHash string
	Role         UserRole
	ImagePath    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepository defines the interface for user data access.
type UserRepository interface {
	// Create creates a new user and sets its ID.
	Create(ctx context.Context, user *User) error

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id int) (*User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update updates an existing user's information.
	Update(ctx context.Context, user *User) error

	// Delete removes a user by their ID.
	Delete(ctx context.Context, id int) error

	// ListWithCount returns paginated list of users.
	// Returns users slice, total count, and error.
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
