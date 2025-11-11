package auth

import "context"

type AuthenticatedUser struct {
	ID   int
	Role string
}

type Auth interface {
	GetAuthUser(ctx context.Context) (AuthenticatedUser, error)
}
