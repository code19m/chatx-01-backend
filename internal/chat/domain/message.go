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
	Create(ctx context.Context, message *Message) error
	GetByID(ctx context.Context, id int) (*Message, error)
	Update(ctx context.Context, message *Message) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context, chatID int, offset, limit int) ([]*Message, int, error)
	GetUnreadCountByChat(ctx context.Context, chatID, userID int) (int, error)
	GetTotalUnreadCount(ctx context.Context, userID int) (int, error)
}
