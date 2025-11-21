package ws

import "time"

// Broadcaster defines the interface for broadcasting WebSocket events.
// This interface allows the use cases to broadcast events without
// depending on the concrete WebSocket implementation.
type Broadcaster interface {
	// BroadcastNewMessage broadcasts a new message event to chat participants.
	BroadcastNewMessage(chatID, messageID, senderID int, content string, sentAt time.Time)

	// BroadcastEditMessage broadcasts a message edit event to chat participants.
	BroadcastEditMessage(chatID, messageID, senderID int, content string, editedAt time.Time)

	// BroadcastDeleteMessage broadcasts a message deletion event to chat participants.
	BroadcastDeleteMessage(chatID, messageID int)

	// BroadcastReadReceipt broadcasts a read receipt event to chat participants.
	BroadcastReadReceipt(chatID, userID, messageID int, readAt time.Time)
}

// hubBroadcaster implements Broadcaster using the Hub.
type hubBroadcaster struct {
	hub *Hub
}

// NewBroadcaster creates a new Broadcaster using the provided Hub.
func NewBroadcaster(hub *Hub) Broadcaster {
	return &hubBroadcaster{hub: hub}
}

func (b *hubBroadcaster) BroadcastNewMessage(chatID, messageID, senderID int, content string, sentAt time.Time) {
	event := &Event{
		Type: EventMessageNew,
		Payload: MessagePayload{
			ID:       messageID,
			ChatID:   chatID,
			SenderID: senderID,
			Content:  content,
			SentAt:   sentAt,
		},
	}
	b.hub.BroadcastToChat(chatID, event, 0) // Include sender
}

func (b *hubBroadcaster) BroadcastEditMessage(chatID, messageID, senderID int, content string, editedAt time.Time) {
	event := &Event{
		Type: EventMessageEdit,
		Payload: MessagePayload{
			ID:       messageID,
			ChatID:   chatID,
			SenderID: senderID,
			Content:  content,
			EditedAt: &editedAt,
		},
	}
	b.hub.BroadcastToChat(chatID, event, 0) // Include sender
}

func (b *hubBroadcaster) BroadcastDeleteMessage(chatID, messageID int) {
	event := &Event{
		Type: EventMessageDelete,
		Payload: MessageDeletePayload{
			ID:     messageID,
			ChatID: chatID,
		},
	}
	b.hub.BroadcastToChat(chatID, event, 0) // Include sender
}

func (b *hubBroadcaster) BroadcastReadReceipt(chatID, userID, messageID int, readAt time.Time) {
	event := &Event{
		Type: EventMessageRead,
		Payload: MessageReadPayload{
			ChatID:    chatID,
			UserID:    userID,
			MessageID: messageID,
			ReadAt:    readAt,
		},
	}
	b.hub.BroadcastToChat(chatID, event, userID) // Exclude the reader
}

// NopBroadcaster is a no-op broadcaster for testing or when WebSocket is disabled.
type NopBroadcaster struct{}

func (NopBroadcaster) BroadcastNewMessage(chatID, messageID, senderID int, content string, sentAt time.Time) {
}
func (NopBroadcaster) BroadcastEditMessage(chatID, messageID, senderID int, content string, editedAt time.Time) {
}
func (NopBroadcaster) BroadcastDeleteMessage(chatID, messageID int) {}
func (NopBroadcaster) BroadcastReadReceipt(chatID, userID, messageID int, readAt time.Time) {}
