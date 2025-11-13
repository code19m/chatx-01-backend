package portal

import (
	"chatx-01-backend/internal/portal/auth"
	"context"
)

// Interface guard.
var _ auth.Auth = (*Portal)(nil)

type Portal struct{}

func (p *Portal) GetAuthUser(ctx context.Context) (auth.AuthenticatedUser, error) {
	return auth.AuthenticatedUser{}, nil
}
