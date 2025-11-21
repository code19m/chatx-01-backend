package ws

import (
	"context"
	"log/slog"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	// clients maps userID to their active client connections.
	// A user can have multiple connections (different devices/tabs).
	clients map[int]map[*Client]struct{}

	// chatSubscriptions maps chatID to userIDs participating.
	// Used for efficient broadcasting to chat participants.
	chatSubscriptions map[int]map[int]struct{}

	// register requests from clients.
	register chan *Client

	// unregister requests from clients.
	unregister chan *Client

	// broadcast channel for events to be sent to specific chats.
	broadcast chan *BroadcastMessage

	// userBroadcast channel for events to be sent to specific users.
	userBroadcast chan *UserBroadcastMessage

	mu     sync.RWMutex
	logger *slog.Logger
}

// BroadcastMessage contains an event to be broadcast to a chat.
type BroadcastMessage struct {
	ChatID    int
	Event     *Event
	ExcludeID int // UserID to exclude from broadcast (e.g., sender)
}

// UserBroadcastMessage contains an event to be sent to a specific user.
type UserBroadcastMessage struct {
	UserID int
	Event  *Event
}

// NewHub creates a new Hub instance.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients:           make(map[int]map[*Client]struct{}),
		chatSubscriptions: make(map[int]map[int]struct{}),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan *BroadcastMessage, 256),
		userBroadcast:     make(chan *UserBroadcastMessage, 256),
		logger:            logger,
	}
}

// Run starts the hub's main event loop.
// This should be run in a separate goroutine.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case msg := <-h.broadcast:
			h.broadcastToChat(msg)

		case msg := <-h.userBroadcast:
			h.broadcastToUser(msg)
		}
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// BroadcastToChat sends an event to all participants of a chat.
func (h *Hub) BroadcastToChat(chatID int, event *Event, excludeUserID int) {
	h.broadcast <- &BroadcastMessage{
		ChatID:    chatID,
		Event:     event,
		ExcludeID: excludeUserID,
	}
}

// BroadcastToUser sends an event to a specific user.
func (h *Hub) BroadcastToUser(userID int, event *Event) {
	h.userBroadcast <- &UserBroadcastMessage{
		UserID: userID,
		Event:  event,
	}
}

// SubscribeToChat adds a user to a chat's subscription list.
func (h *Hub) SubscribeToChat(chatID, userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.chatSubscriptions[chatID] == nil {
		h.chatSubscriptions[chatID] = make(map[int]struct{})
	}
	h.chatSubscriptions[chatID][userID] = struct{}{}
}

// UnsubscribeFromChat removes a user from a chat's subscription list.
func (h *Hub) UnsubscribeFromChat(chatID, userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if users, ok := h.chatSubscriptions[chatID]; ok {
		delete(users, userID)
		if len(users) == 0 {
			delete(h.chatSubscriptions, chatID)
		}
	}
}

// IsUserOnline checks if a user has any active connections.
func (h *Hub) IsUserOnline(userID int) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// GetOnlineUsers returns a list of user IDs that are currently online.
func (h *Hub) GetOnlineUsers(userIDs []int) []int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	online := make([]int, 0)
	for _, userID := range userIDs {
		if clients, ok := h.clients[userID]; ok && len(clients) > 0 {
			online = append(online, userID)
		}
	}
	return online
}

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userID := client.UserID()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*Client]struct{})
	}
	h.clients[userID][client] = struct{}{}

	// Subscribe to all user's chats
	for _, chatID := range client.ChatIDs() {
		if h.chatSubscriptions[chatID] == nil {
			h.chatSubscriptions[chatID] = make(map[int]struct{})
		}
		h.chatSubscriptions[chatID][userID] = struct{}{}
	}

	h.logger.Info("client registered",
		"user_id", userID,
		"total_connections", len(h.clients[userID]),
	)
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	userID := client.UserID()
	if clients, ok := h.clients[userID]; ok {
		if _, exists := clients[client]; exists {
			delete(clients, client)
			client.Close()

			if len(clients) == 0 {
				delete(h.clients, userID)
				// Clean up chat subscriptions for this user
				for chatID, users := range h.chatSubscriptions {
					delete(users, userID)
					if len(users) == 0 {
						delete(h.chatSubscriptions, chatID)
					}
				}
			}

			h.logger.Info("client unregistered",
				"user_id", userID,
				"remaining_connections", len(h.clients[userID]),
			)
		}
	}
}

func (h *Hub) broadcastToChat(msg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users, ok := h.chatSubscriptions[msg.ChatID]
	if !ok {
		return
	}

	for userID := range users {
		if userID == msg.ExcludeID {
			continue
		}
		h.sendToUser(userID, msg.Event)
	}
}

func (h *Hub) broadcastToUser(msg *UserBroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.sendToUser(msg.UserID, msg.Event)
}

// sendToUser sends an event to all connections of a user.
// Must be called with read lock held.
func (h *Hub) sendToUser(userID int, event *Event) {
	clients, ok := h.clients[userID]
	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.send <- event:
		default:
			// Client's send buffer is full, skip this message
			h.logger.Warn("client send buffer full, dropping message",
				"user_id", userID,
			)
		}
	}
}

func (h *Hub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for userID, clients := range h.clients {
		for client := range clients {
			client.Close()
		}
		delete(h.clients, userID)
	}

	h.chatSubscriptions = make(map[int]map[int]struct{})
	h.logger.Info("hub shutdown complete")
}
