package infra

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"chatx-01-backend/internal/chat/domain"
	"chatx-01-backend/pkg/errs"
	"chatx-01-backend/pkg/pg"
)

type PgChatRepo struct {
	pool *pgxpool.Pool
}

func NewPgChatRepo(pool *pgxpool.Pool) *PgChatRepo {
	return &PgChatRepo{
		pool: pool,
	}
}

func (r *PgChatRepo) Create(ctx context.Context, chat *domain.Chat) error {
	const op = "pgchat.Create"

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
		return pg.WrapRepoError(op, err)
	}

	return nil
}

func (r *PgChatRepo) GetByID(ctx context.Context, id int) (*domain.Chat, error) {
	const op = "pgchat.GetByID"

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
		return nil, pg.WrapRepoError(op, err)
	}

	return chat, nil
}

func (r *PgChatRepo) GetDMByParticipants(ctx context.Context, userID1, userID2 int) (*domain.Chat, error) {
	const op = "pgchat.GetDMByParticipants"

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
		return nil, pg.WrapRepoError(op, err)
	}

	return chat, nil
}

func (r *PgChatRepo) GetDMsListByUser(ctx context.Context, userID int, offset, limit int) ([]domain.Chat, int, error) {
	const op = "pgchat.GetDMsListByUser"

	var totalCount int
	countQuery := `
		SELECT COUNT(DISTINCT c.id)
		FROM chats c
		INNER JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = $1 AND c.type = $2`

	err := r.pool.QueryRow(ctx, countQuery, userID, domain.ChatTypeDirect).Scan(&totalCount)
	if err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}

	query := `
		SELECT c.id, c.type, c.name, c.creator_id, c.created_at
		FROM chats c
		INNER JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = $1 AND c.type = $2
		ORDER BY c.created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, userID, domain.ChatTypeDirect, limit, offset)
	if err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}
	defer rows.Close()

	chats := make([]domain.Chat, 0)
	for rows.Next() {
		chat := domain.Chat{}
		err := rows.Scan(
			&chat.ID,
			&chat.Type,
			&chat.Name,
			&chat.CreatorID,
			&chat.CreatedAt,
		)
		if err != nil {
			return nil, 0, pg.WrapRepoError(op, err)
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}

	return chats, totalCount, nil
}

func (r *PgChatRepo) GetGroupsListByUser(
	ctx context.Context,
	userID int,
	offset, limit int,
) ([]domain.Chat, int, error) {
	const op = "pgchat.GetGroupsListByUser"

	var totalCount int
	countQuery := `
		SELECT COUNT(DISTINCT c.id)
		FROM chats c
		INNER JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = $1 AND c.type = $2`

	err := r.pool.QueryRow(ctx, countQuery, userID, domain.ChatTypeGroup).Scan(&totalCount)
	if err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}

	query := `
		SELECT c.id, c.type, c.name, c.creator_id, c.created_at
		FROM chats c
		INNER JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = $1 AND c.type = $2
		ORDER BY c.created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, userID, domain.ChatTypeGroup, limit, offset)
	if err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}
	defer rows.Close()

	chats := make([]domain.Chat, 0)
	for rows.Next() {
		chat := domain.Chat{}
		err := rows.Scan(
			&chat.ID,
			&chat.Type,
			&chat.Name,
			&chat.CreatorID,
			&chat.CreatedAt,
		)
		if err != nil {
			return nil, 0, pg.WrapRepoError(op, err)
		}
		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, pg.WrapRepoError(op, err)
	}

	return chats, totalCount, nil
}

func (r *PgChatRepo) AddParticipant(ctx context.Context, participant *domain.ChatParticipant) error {
	const op = "pgchat.AddParticipant"

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
		return pg.WrapRepoError(op, err)
	}

	return nil
}

func (r *PgChatRepo) RemoveParticipant(ctx context.Context, chatID, userID int) error {
	const op = "pgchat.RemoveParticipant"

	query := `DELETE FROM chat_participants WHERE chat_id = $1 AND user_id = $2`

	result, err := r.pool.Exec(ctx, query, chatID, userID)
	if err != nil {
		return pg.WrapRepoError(op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errs.Wrap(op, errors.New("no rows affected"))
	}

	return nil
}

func (r *PgChatRepo) GetParticipants(ctx context.Context, chatID int) ([]domain.ChatParticipant, error) {
	const op = "pgchat.GetParticipants"

	query := `
		SELECT chat_id, user_id, joined_at, last_read_message_id, last_read_at
		FROM chat_participants
		WHERE chat_id = $1
		ORDER BY joined_at ASC`

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, pg.WrapRepoError(op, err)
	}
	defer rows.Close()

	participants := make([]domain.ChatParticipant, 0)
	for rows.Next() {
		participant := domain.ChatParticipant{}
		err := rows.Scan(
			&participant.ChatID,
			&participant.UserID,
			&participant.JoinedAt,
			&participant.LastReadMessageID,
			&participant.LastReadAt,
		)
		if err != nil {
			return nil, pg.WrapRepoError(op, err)
		}
		participants = append(participants, participant)
	}

	if err := rows.Err(); err != nil {
		return nil, pg.WrapRepoError(op, err)
	}

	return participants, nil
}

func (r *PgChatRepo) IsParticipant(ctx context.Context, chatID, userID int) (bool, error) {
	const op = "pgchat.IsParticipant"

	query := `SELECT EXISTS(SELECT 1 FROM chat_participants WHERE chat_id = $1 AND user_id = $2)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, chatID, userID).Scan(&exists)
	if err != nil {
		return false, pg.WrapRepoError(op, err)
	}

	return exists, nil
}

func (r *PgChatRepo) UpdateLastRead(ctx context.Context, chatID, userID, messageID int) error {
	const op = "pgchat.UpdateLastRead"

	query := `
		UPDATE chat_participants
		SET last_read_message_id = $1, last_read_at = NOW()
		WHERE chat_id = $2 AND user_id = $3`

	result, err := r.pool.Exec(ctx, query, messageID, chatID, userID)
	if err != nil {
		return pg.WrapRepoError(op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errs.Wrap(op, errors.New("no rows affected"))
	}

	return nil
}
