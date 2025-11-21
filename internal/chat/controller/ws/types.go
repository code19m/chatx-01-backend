package ws

import "time"

// EventType represents the type of WebSocket event.
type EventType string

const (
	// Message events
	EventMessageNew    EventType = "message.new"
	EventMessageEdit   EventType = "message.edit"
	EventMessageDelete EventType = "message.delete"
	EventMessageRead   EventType = "message.read"

	// Typing events
	EventTypingStart EventType = "typing.start"
	EventTypingStop  EventType = "typing.stop"

	// Presence events
	EventPresenceOnline  EventType = "presence.online"
	EventPresenceOffline EventType = "presence.offline"

	// Error events
	EventError EventType = "error"
)

// Event is the base WebSocket message envelope.
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// MessagePayload contains message data for message events.
type MessagePayload struct {
	ID       int        `json:"id"`
	ChatID   int        `json:"chat_id"`
	SenderID int        `json:"sender_id"`
	Content  string     `json:"content,omitempty"`
	SentAt   time.Time  `json:"sent_at,omitempty"`
	EditedAt *time.Time `json:"edited_at,omitempty"`
}

// MessageDeletePayload contains data for message deletion events.
type MessageDeletePayload struct {
	ID     int `json:"id"`
	ChatID int `json:"chat_id"`
}

// MessageReadPayload contains data for read receipt events.
type MessageReadPayload struct {
	ChatID    int       `json:"chat_id"`
	UserID    int       `json:"user_id"`
	MessageID int       `json:"message_id"`
	ReadAt    time.Time `json:"read_at"`
}

// TypingPayload contains data for typing indicator events.
type TypingPayload struct {
	ChatID int `json:"chat_id"`
	UserID int `json:"user_id"`
}

// PresencePayload contains data for presence events.
type PresencePayload struct {
	UserID   int        `json:"user_id"`
	Online   bool       `json:"online"`
	LastSeen *time.Time `json:"last_seen,omitempty"`
}

// ErrorPayload contains error information.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ClientMessage represents a message sent from client to server.
type ClientMessage struct {
	Type    EventType       `json:"type"`
	Payload ClientPayload   `json:"payload"`
}

// ClientPayload is the payload for client-sent messages.
type ClientPayload struct {
	ChatID int `json:"chat_id,omitempty"`
}
