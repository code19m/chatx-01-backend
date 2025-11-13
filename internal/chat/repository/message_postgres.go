package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"chatx-01-backend/internal/chat/domain"
)

// MessagePostgresRepository implements domain.MessageRepository using PostgreSQL
type MessagePostgresRepository struct {
	pool *pgxpool.Pool
}

// NewMessagePostgresRepository creates a new PostgreSQL message repository
func NewMessagePostgresRepository(pool *pgxpool.Pool) *MessagePostgresRepository {
	return &MessagePostgresRepository{
		pool: pool,
	}
}

// Create creates a new message in the database
func (r *MessagePostgresRepository) Create(ctx context.Context, message *domain.Message) error {
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
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetByID retrieves a message by its ID
func (r *MessagePostgresRepository) GetByID(ctx context.Context, id int) (*domain.Message, error) {
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("message not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get message by id: %w", err)
	}

	return message, nil
}

// Update updates an existing message in the database
func (r *MessagePostgresRepository) Update(ctx context.Context, message *domain.Message) error {
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
		return fmt.Errorf("failed to update message: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// Delete deletes a message from the database
func (r *MessagePostgresRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM messages WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("message not found")
	}

	return nil
}

// List retrieves a paginated list of messages for a chat with total count
func (r *MessagePostgresRepository) List(ctx context.Context, chatID int, offset, limit int) ([]*domain.Message, int, error) {
	// Get total count
	var totalCount int
	countQuery := `SELECT COUNT(*) FROM messages WHERE chat_id = $1`
	err := r.pool.QueryRow(ctx, countQuery, chatID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	// Get paginated messages
	query := `
		SELECT id, chat_id, sender_id, content, sent_at, edited_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY sent_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, chatID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list messages: %w", err)
	}
	defer rows.Close()

	messages := make([]*domain.Message, 0)
	for rows.Next() {
		message := &domain.Message{}
		err := rows.Scan(
			&message.ID,
			&message.ChatID,
			&message.SenderID,
			&message.Content,
			&message.SentAt,
			&message.EditedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, totalCount, nil
}

// GetUnreadCountByChat returns the count of unread messages in a specific chat for a user
func (r *MessagePostgresRepository) GetUnreadCountByChat(ctx context.Context, chatID, userID int) (int, error) {
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
		return 0, fmt.Errorf("failed to get unread count by chat: %w", err)
	}

	return count, nil
}

// GetTotalUnreadCount returns the total count of unread messages across all chats for a user
func (r *MessagePostgresRepository) GetTotalUnreadCount(ctx context.Context, userID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		INNER JOIN chat_participants cp ON m.chat_id = cp.chat_id AND cp.user_id = $1
		WHERE m.sender_id != $1
		AND (cp.last_read_message_id IS NULL OR m.id > cp.last_read_message_id)`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total unread count: %w", err)
	}

	return count, nil
}
