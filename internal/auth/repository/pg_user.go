package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"chatx-01-backend/internal/auth/domain"
	"chatx-01-backend/pkg/errs"
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
	const op = "pguserrepo.Create"

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

	// TODO: handle conflict
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.Wrap(op, errs.ErrNotFound)
	}
	if err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (r *PgUserRepo) GetByID(ctx context.Context, id int) (*domain.User, error) {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

func (r *PgUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *PgUserRepo) Update(ctx context.Context, user *domain.User) error {
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
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *PgUserRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *PgUserRepo) List(ctx context.Context, offset, limit int) ([]*domain.User, int, error) {
	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.pool.QueryRow(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated users
	query := `
		SELECT id, email, username, password_hash, role, image_path, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
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
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, totalCount, nil
}
