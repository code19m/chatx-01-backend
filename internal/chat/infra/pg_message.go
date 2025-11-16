package infra

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"chatx-01-backend/internal/chat/domain"
	"chatx-01-backend/pkg/errs"
	"chatx-01-backend/pkg/pg"
)

type PgMessageRepo struct {
	pool *pgxpool.Pool
}

func NewPgMessageRepo(pool *pgxpool.Pool) *PgMessageRepo {
	return &PgMessageRepo{
		pool: pool,
	}
}

func (r *PgMessageRepo) Create(ctx context.Context, message *domain.Message) error {
	const op = "pgmessage.Create"

	query := `
		INSERT INTO messages (chat_id, sender_id, content, sent_at, edited_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.pool.QueryRow(
		ctx,
		query,
		message.ChatID,
		message.SenderID,
		message.Content,
		message.SentAt,
		message.EditedAt,
	).Scan(&message.ID)
	if err != nil {
		return pg.WrapRepoError(op, err)
	}

	return nil
}

func (r *PgMessageRepo) GetByID(ctx context.Context, id int) (*domain.Message, error) {
	const op = "pgmessage.GetByID"

	query := `
		SELECT id, chat_id, sender_id, content, sent_at, edited_at
		FROM messages
		WHERE id = $1`

	message := &domain.Message{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&message.ID,
		&message.ChatID,
		&message.SenderID,
		&message.Content,
		&message.SentAt,
		&message.EditedAt,
	)
	if err != nil {
		return nil, pg.WrapRepoError(op, err)
	}

	return message, nil
}

func (r *PgMessageRepo) Update(ctx context.Context, message *domain.Message) error {
	const op = "pgmessage.Update"

	query := `
		UPDATE messages
		SET content = $1, edited_at = $2
		WHERE id = $3`

	result, err := r.pool.Exec(
		ctx,
		query,
		message.Content,
		message.EditedAt,
		message.ID,
	)
	if err != nil {
		return pg.WrapRepoError(op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errs.Wrap(op, errors.New("no rows affected"))
	}

	return nil
}

func (r *PgMessageRepo) Delete(ctx context.Context, id int) error {
	const op = "pgmessage.Delete"

	query := `DELETE FROM messages WHERE id = $1`

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

func (r *PgMessageRepo) ListWithCount(
	ctx context.Context,
	chatID int,
	offset, limit int,
) ([]domain.Message, int, error) {
	const op = "pgmessage.List"

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM messages WHERE chat_id = $1`
	err := r.pool.QueryRow(ctx, countQuery, chatID).Scan(&totalCount)
	if err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}

	query := `
		SELECT id, chat_id, sender_id, content, sent_at, edited_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY sent_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, chatID, limit, offset)
	if err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}
	defer rows.Close()

	messages := make([]domain.Message, 0)
	for rows.Next() {
		message := domain.Message{}
		err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.Content,
			&message.SentAt,
			&message.EditedAt,
		)
		if err != nil {
			return nil, 0, pg.WrapRepoError(op, err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}

	return messages, totalCount, nil
}

func (r *PgMessageRepo) GetLastMessage(ctx context.Context, chatID int) (*domain.Message, error) {
	const op = "pgmessage.GetLastMessage"

	query := `
		SELECT id, chat_id, sender_id, content, sent_at, edited_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY sent_at DESC
		LIMIT 1`

	message := &domain.Message{}
	err := r.pool.QueryRow(ctx, query, chatID).Scan(
		&message.ID,
		&message.ChatID,
		&message.SenderID,
		&message.Content,
		&message.SentAt,
		&message.EditedAt,
	)
	if err != nil {
		return nil, pg.WrapRepoError(op, err)
	}

	return message, nil
}

func (r *PgMessageRepo) GetUnreadCountByChat(ctx context.Context, chatID, userID int) (int, error) {
	const op = "pgmessage.GetUnreadCountByChat"

	query := `
		SELECT COUNT(*)
		FROM messages m
		LEFT JOIN chat_participants cp ON m.chat_id = cp.chat_id AND cp.user_id = $2
		WHERE m.chat_id = $1
		AND m.sender_id != $2
		AND (cp.last_read_message_id IS NULL OR m.id > cp.last_read_message_id)`

	var count int
	err := r.pool.QueryRow(ctx, query, chatID, userID).Scan(&count)
	if err != nil {
		return 0, pg.WrapRepoError(op, err)
	}

	return count, nil
}

func (r *PgMessageRepo) GetTotalUnreadCount(ctx context.Context, userID int) (int, error) {
	const op = "pgmessage.GetTotalUnreadCount"

	query := `
		SELECT COUNT(*)
		FROM messages m
		INNER JOIN chat_participants cp ON m.chat_id = cp.chat_id AND cp.user_id = $1
		WHERE m.sender_id != $1
		AND (cp.last_read_message_id IS NULL OR m.id > cp.last_read_message_id)`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, pg.WrapRepoError(op, err)
	}

	return count, nil
}
