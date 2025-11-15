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
	// Create creates a new chat and sets its ID.
	Create(ctx context.Context, chat *Chat) error

	// GetByID retrieves a chat by its ID.
	GetByID(ctx context.Context, id int) (*Chat, error)

	// GetDMByParticipants finds a direct message chat between two users.
	GetDMByParticipants(ctx context.Context, userID1, userID2 int) (*Chat, error)

	// GetDMsListByUser returns paginated list of direct message chats for a user.
	// Returns chats slice, total count, and error.
	GetDMsListByUser(ctx context.Context, userID int, offset, limit int) ([]*Chat, int, error)

	// GetGroupsListByUser returns paginated list of group chats for a user.
	// Returns chats slice, total count, and error.
	GetGroupsListByUser(ctx context.Context, userID int, offset, limit int) ([]*Chat, int, error)

	// AddParticipant adds a user to a chat.
	AddParticipant(ctx context.Context, participant *ChatParticipant) error

	// RemoveParticipant removes a user from a chat.
	RemoveParticipant(ctx context.Context, chatID, userID int) error

	// GetParticipants retrieves all participants of a chat.
	GetParticipants(ctx context.Context, chatID int) ([]*ChatParticipant, error)

	// IsParticipant checks if a user is a participant of a chat.
	IsParticipant(ctx context.Context, chatID, userID int) (bool, error)

	// UpdateLastRead updates the last read message for a participant.
	UpdateLastRead(ctx context.Context, chatID, userID, messageID int) error
}
