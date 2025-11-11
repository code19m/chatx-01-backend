package domain

import (
	"context"
	"time"
)

type ChatType string

const (
	ChatTypeDirect ChatType = "direct"
	ChatTypeGroup  ChatType = "group"
)

type Chat struct {
	ID        int
	Type      ChatType
	Name      string // Empty for direct chats
	CreatorID int
	CreatedAt time.Time
}

type ChatParticipant struct {
	ChatID            int
	UserID            int
	JoinedAt          time.Time
	LastReadMessageID *int
	LastReadAt        *time.Time // DENORMALIZED
}

// ChatRepository defines the interface for chat data access.
type ChatRepository interface {
	Create(ctx context.Context, chat *Chat) error
	GetByID(ctx context.Context, id int) (*Chat, error)
	GetDMByParticipants(ctx context.Context, userID1, userID2 int) (*Chat, error)
	GetDMsListByUser(ctx context.Context, userID int, offset, limit int) ([]*Chat, int, error)
	GetGroupsListByUser(ctx context.Context, userID int, offset, limit int) ([]*Chat, int, error)
	AddParticipant(ctx context.Context, participant *ChatParticipant) error
	RemoveParticipant(ctx context.Context, chatID, userID int) error
	GetParticipants(ctx context.Context, chatID int) ([]*ChatParticipant, error)
	IsParticipant(ctx context.Context, chatID, userID int) (bool, error)
	UpdateLastRead(ctx context.Context, chatID, userID, messageID int) error
}
