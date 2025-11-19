package infra

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"chatx-01-backend/internal/auth/domain"
	"chatx-01-backend/pkg/errs"
	"chatx-01-backend/pkg/pg"
)

type PgUserRepo struct {
	pool *pgxpool.Pool
}

func NewPgUserRepo(pool *pgxpool.Pool) *PgUserRepo {
	return &PgUserRepo{
		pool: pool,
	}
}

func (r *PgUserRepo) Create(ctx context.Context, user *domain.User) error {
	const op = "pguser.Create"

	query := `
		INSERT INTO users (email, username, password_hash, role, image_path, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := r.pool.QueryRow(
		ctx,
		query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.ImagePath,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
	if err != nil {
		return pg.WrapRepoError(op, err)
	}

	return nil
}

func (r *PgUserRepo) GetByID(ctx context.Context, id int) (*domain.User, error) {
	const op = "pguser.GetByID"

	query := `
		SELECT id, email, username, password_hash, role, image_path, created_at, updated_at
		FROM users
		WHERE id = $1`

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.ImagePath,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, pg.WrapRepoError(op, err)
	}

	return user, nil
}

func (r *PgUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const op = "pguser.GetByEmail"

	query := `
		SELECT id, email, username, password_hash, role, image_path, created_at, updated_at
		FROM users
		WHERE email = $1`

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.ImagePath,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, pg.WrapRepoError(op, err)
	}

	return user, nil
}

func (r *PgUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	const op = "pguser.GetByUsername"

	query := `
		SELECT id, email, username, password_hash, role, image_path, created_at, updated_at
		FROM users
		WHERE username = $1`

	user := &domain.User{}
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.ImagePath,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, pg.WrapRepoError(op, err)
	}

	return user, nil
}

func (r *PgUserRepo) Update(ctx context.Context, user *domain.User) error {
	const op = "pguser.Update"

	query := `
		UPDATE users
		SET email = $1, username = $2, password_hash = $3, role = $4, image_path = $5, updated_at = $6
		WHERE id = $7`

	result, err := r.pool.Exec(
		ctx,
		query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.ImagePath,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return pg.WrapRepoError(op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errs.Wrap(op, errors.New("no rofs affected"))
	}

	return nil
}

func (r *PgUserRepo) Delete(ctx context.Context, id int) error {
	const op = "pguser.Delete"

	query := `DELETE FROM users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return pg.WrapRepoError(op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errs.Wrap(op, errors.New("no rows affected"))
	}

	return nil
}

func (r *PgUserRepo) ListWithCount(ctx context.Context, offset, limit int) ([]*domain.User, int, error) {
	const op = "pguser.ListWithCount"

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.pool.QueryRow(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}

	query := `
		SELECT id, email, username, password_hash, role, image_path, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&user.ImagePath,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, pg.WrapRepoError(op, err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}

	return users, totalCount, nil
}

func (r *PgUserRepo) SearchByUsernameWithCount(ctx context.Context, username string, offset, limit int) ([]*domain.User, int, error) {
	const op = "pguser.SearchByUsernameWithCount"

	searchPattern := "%" + username + "%"

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM users WHERE username ILIKE $1`
	err := r.pool.QueryRow(ctx, countQuery, searchPattern).Scan(&totalCount)
	if err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}

	query := `
		SELECT id, email, username, password_hash, role, image_path, created_at, updated_at
		FROM users
		WHERE username ILIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&user.ImagePath,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, pg.WrapRepoError(op, err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}

	return users, totalCount, nil
}
