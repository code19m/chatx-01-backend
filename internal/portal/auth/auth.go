package auth

import (
	"context"
	"net/http"
)

type AuthenticatedUser struct {
	ID   int
	Role string
}

type User struct {
	ID        int
	Email     string
	Username  string
	Role      string
	ImagePath *string
}

type Portal interface {
	// GetAuthUser retrieves the authenticated user from context.
	GetAuthUser(ctx context.Context) (AuthenticatedUser, error)

	// SetAuthUser sets the authenticated user in context.
	SetAuthUser(ctx context.Context, au AuthenticatedUser) context.Context

	// GetUserByID retrieves a user by their ID.
	GetUserByID(ctx context.Context, id int) (*User, error)

	// GetUsersByIDs retrieves multiple users by their IDs.
	GetUsersByIDs(ctx context.Context, ids []int) ([]*User, error)

	// UserExists checks if a user exists by ID.
	UserExists(ctx context.Context, id int) (bool, error)

	// RequireAuth returns a middleware that checks if the user is authenticated.
	RequireAuth() func(next http.Handler) http.Handler

	// RequireAdmin creates a middleware that checks if the user is authenticated and has admin role.
	RequireAdmin() func(next http.Handler) http.Handler
}
