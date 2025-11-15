package portal

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"chatx-01-backend/internal/auth/domain"
	"chatx-01-backend/internal/portal/auth"
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
	userRepo domain.UserRepository
}

func New(userRepo domain.UserRepository) *Portal {
	return &Portal{
		userRepo: userRepo,
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
