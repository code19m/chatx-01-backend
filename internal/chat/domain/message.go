package domain

import (
	"context"
	"time"
)

type Message struct {
	ID       int
	ChatID   int
	SenderID int
	Content  string
	SentAt   time.Time
	EditedAt *time.Time
}

// MessageRepository defines the interface for message data access.
type MessageRepository interface {
	// Create creates a new message and sets its ID.
	Create(ctx context.Context, message *Message) error

	// GetByID retrieves a message by its ID.
	GetByID(ctx context.Context, id int) (*Message, error)

	// Update updates an existing message's content and edited timestamp.
	Update(ctx context.Context, message *Message) error

	// Delete removes a message by its ID.
	Delete(ctx context.Context, id int) error

	// List returns paginated list of messages in a chat.
	// Returns messages slice, total count, and error.
	List(ctx context.Context, chatID int, offset, limit int) ([]*Message, int, error)

	// GetLastMessage returns the most recent message in a chat, or nil if no messages exist.
	GetLastMessage(ctx context.Context, chatID int) (*Message, error)

	// GetUnreadCountByChat returns the count of unread messages in a specific chat for a user.
	GetUnreadCountByChat(ctx context.Context, chatID, userID int) (int, error)

	// GetTotalUnreadCount returns the total count of unread messages across all chats for a user.
	GetTotalUnreadCount(ctx context.Context, userID int) (int, error)
}
