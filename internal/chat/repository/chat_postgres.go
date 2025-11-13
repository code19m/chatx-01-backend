package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"chatx-01-backend/internal/chat/domain"
)

// ChatPostgresRepository implements domain.ChatRepository using PostgreSQL
type ChatPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewChatPostgresRepository creates a new PostgreSQL chat repository
func NewChatPostgresRepository(pool *pgxpool.Pool) *ChatPostgresRepository {
	return &ChatPostgresRepository{
		pool: pool,
	}
}

// Create creates a new chat in the database
func (r *ChatPostgresRepository) Create(ctx context.Context, chat *domain.Chat) error {
	query := `
		INSERT INTO chats (type, name, creator_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	err := r.pool.QueryRow(
		ctx,
		query,
		chat.Type,
		chat.Name,
		chat.CreatorID,
		chat.CreatedAt,
	).Scan(&chat.ID)

	if err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}

	return nil
}

// GetByID retrieves a chat by its ID
func (r *ChatPostgresRepository) GetByID(ctx context.Context, id int) (*domain.Chat, error) {
	query := `
		SELECT id, type, name, creator_id, created_at
		FROM chats
		WHERE id = $1`

	chat := &domain.Chat{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&chat.ID,
		&chat.Type,
		&chat.Name,
		&chat.CreatorID,
		&chat.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("chat not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get chat by id: %w", err)
	}

	return chat, nil
}

// GetDMByParticipants retrieves a direct message chat between two users
func (r *ChatPostgresRepository) GetDMByParticipants(ctx context.Context, userID1, userID2 int) (*domain.Chat, error) {
	query := `
		SELECT c.id, c.type, c.name, c.creator_id, c.created_at
		FROM chats c
		INNER JOIN chat_participants cp1 ON c.id = cp1.chat_id AND cp1.user_id = $1
		INNER JOIN chat_participants cp2 ON c.id = cp2.chat_id AND cp2.user_id = $2
		WHERE c.type = $3`

	chat := &domain.Chat{}
	err := r.pool.QueryRow(ctx, query, userID1, userID2, domain.ChatTypeDirect).Scan(
		&chat.ID,
		&chat.Type,
		&chat.Name,
		&chat.CreatorID,
		&chat.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("dm not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get dm by participants: %w", err)
	}

	return chat, nil
}

// GetDMsListByUser retrieves all direct message chats for a user with pagination
func (r *ChatPostgresRepository) GetDMsListByUser(ctx context.Context, userID int, offset, limit int) ([]*domain.Chat, int, error) {
	// Get total count
	var totalCount int
	countQuery := `
		SELECT COUNT(DISTINCT c.id)
		FROM chats c
		INNER JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = $1 AND c.type = $2`

	err := r.pool.QueryRow(ctx, countQuery, userID, domain.ChatTypeDirect).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count dms: %w", err)
	}

	// Get paginated chats
	query := `
		SELECT c.id, c.type, c.name, c.creator_id, c.created_at
		FROM chats c
		INNER JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = $1 AND c.type = $2
		ORDER BY c.created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, userID, domain.ChatTypeDirect, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list dms: %w", err)
	}
	defer rows.Close()

	chats := make([]*domain.Chat, 0)
	for rows.Next() {
		chat := &domain.Chat{}
		err := rows.Scan(
			&chat.ID,
			&chat.Type,
			&chat.Name,
			&chat.CreatorID,
			&chat.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan chat: %w", err)
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating chats: %w", err)
	}

	return chats, totalCount, nil
}

// GetGroupsListByUser retrieves all group chats for a user with pagination
func (r *ChatPostgresRepository) GetGroupsListByUser(ctx context.Context, userID int, offset, limit int) ([]*domain.Chat, int, error) {
	// Get total count
	var totalCount int
	countQuery := `
		SELECT COUNT(DISTINCT c.id)
		FROM chats c
		INNER JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = $1 AND c.type = $2`

	err := r.pool.QueryRow(ctx, countQuery, userID, domain.ChatTypeGroup).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count groups: %w", err)
	}

	// Get paginated chats
	query := `
		SELECT c.id, c.type, c.name, c.creator_id, c.created_at
		FROM chats c
		INNER JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = $1 AND c.type = $2
		ORDER BY c.created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, userID, domain.ChatTypeGroup, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	chats := make([]*domain.Chat, 0)
	for rows.Next() {
		chat := &domain.Chat{}
		err := rows.Scan(
			&chat.ID,
			&chat.Type,
			&chat.Name,
			&chat.CreatorID,
			&chat.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan chat: %w", err)
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating chats: %w", err)
	}

	return chats, totalCount, nil
}

// AddParticipant adds a participant to a chat
func (r *ChatPostgresRepository) AddParticipant(ctx context.Context, participant *domain.ChatParticipant) error {
	query := `
		INSERT INTO chat_participants (chat_id, user_id, joined_at, last_read_message_id, last_read_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.pool.Exec(
		ctx,
		query,
		participant.ChatID,
		participant.UserID,
		participant.JoinedAt,
		participant.LastReadMessageID,
		participant.LastReadAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// RemoveParticipant removes a participant from a chat
func (r *ChatPostgresRepository) RemoveParticipant(ctx context.Context, chatID, userID int) error {
	query := `DELETE FROM chat_participants WHERE chat_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, chatID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}

// GetParticipants retrieves all participants of a chat
func (r *ChatPostgresRepository) GetParticipants(ctx context.Context, chatID int) ([]*domain.ChatParticipant, error) {
	query := `
		SELECT chat_id, user_id, joined_at, last_read_message_id, last_read_at
		FROM chat_participants
		WHERE chat_id = $1
		ORDER BY joined_at ASC`

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	defer rows.Close()

	participants := make([]*domain.ChatParticipant, 0)
	for rows.Next() {
		participant := &domain.ChatParticipant{}
		err := rows.Scan(
			&participant.ChatID,
			&participant.UserID,
			&participant.JoinedAt,
			&participant.LastReadMessageID,
			&participant.LastReadAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}
		participants = append(participants, participant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating participants: %w", err)
	}

	return participants, nil
}

// IsParticipant checks if a user is a participant of a chat
func (r *ChatPostgresRepository) IsParticipant(ctx context.Context, chatID, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM chat_participants WHERE chat_id = $1 AND user_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, chatID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check participant: %w", err)
	}

	return exists, nil
}

// UpdateLastRead updates the last read message for a participant
func (r *ChatPostgresRepository) UpdateLastRead(ctx context.Context, chatID, userID, messageID int) error {
	query := `
		UPDATE chat_participants
		SET last_read_message_id = $1, last_read_at = NOW()
		WHERE chat_id = $2 AND user_id = $3`

	result, err := r.pool.Exec(ctx, query, messageID, chatID, userID)
	if err != nil {
		return fmt.Errorf("failed to update last read: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("participant not found")
	}

	return nil
}
