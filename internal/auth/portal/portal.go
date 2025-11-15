package portal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"chatx-01-backend/internal/auth/domain"
	"chatx-01-backend/internal/portal/auth"
	"chatx-01-backend/pkg/token"
)

const (
	authUserKey = "authenticated_user"
)

var (
	errNoAuthUser = errors.New("no authenticated user found in context")

	// Interface guard.
	_ auth.Portal = (*Portal)(nil)
)

type Portal struct {
	userRepo       domain.UserRepository
	tokenGenerator token.Generator
}

func New(
	userRepo domain.UserRepository,
	tokenGenerator token.Generator,
) *Portal {
	return &Portal{
		userRepo:       userRepo,
		tokenGenerator: tokenGenerator,
	}
}

func (p *Portal) GetAuthUser(ctx context.Context) (auth.AuthenticatedUser, error) {
	au, ok := ctx.Value(authUserKey).(auth.AuthenticatedUser)
	if !ok {
		return auth.AuthenticatedUser{}, errNoAuthUser
	}
	return au, nil
}

func (p *Portal) GetUserByID(ctx context.Context, id int) (*auth.User, error) {
	u, err := p.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &auth.User{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Role:      u.Role.String(),
		ImagePath: u.ImagePath,
	}, nil
}

func (p *Portal) GetUsersByIDs(ctx context.Context, ids []int) ([]*auth.User, error) {
	if len(ids) == 0 {
		return []*auth.User{}, nil
	}

	users := make([]*auth.User, 0, len(ids))
	for _, id := range ids {
		u, err := p.userRepo.GetByID(ctx, id)
		if err != nil {
			// Skip users that don't exist
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, err
		}

		users = append(users, &auth.User{
			ID:        u.ID,
			Email:     u.Email,
			Username:  u.Username,
			Role:      u.Role.String(),
			ImagePath: u.ImagePath,
		})
	}

	return users, nil
}

func (p *Portal) UserExists(ctx context.Context, id int) (bool, error) {
	_, err := p.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *Portal) RequireAuth() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "unauthorized: missing authorization header", http.StatusUnauthorized)
				return
			}

			// Check for Bearer prefix
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "unauthorized: invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := p.tokenGenerator.Validate(tokenString)
			if err != nil {
				http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			// Check token type (should be access token)
			if claims.Type != string(token.TokenTypeAccess) {
				http.Error(w, "unauthorized: invalid token type", http.StatusUnauthorized)
				return
			}

			// Add authenticated user to context
			authUser := auth.AuthenticatedUser{
				ID:   claims.UserID,
				Role: claims.Role,
			}

			ctx := context.WithValue(r.Context(), authUserKey, authUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (p *Portal) RequireAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p.RequireAuth()(next).ServeHTTP(w, r)

			authUser, err := p.GetAuthUser(r.Context())
			if err != nil {
				http.Error(w, fmt.Sprintf("could not get authenticated user: %v", err), http.StatusInternalServerError)
				return
			}

			if authUser.Role != domain.RoleAdmin.String() {
				http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
