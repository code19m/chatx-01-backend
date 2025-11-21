package ws

import (
	"context"
	"log/slog"
	"net/http"

	"nhooyr.io/websocket"

	"chatx-01-backend/internal/chat/domain"
	"chatx-01-backend/internal/portal/auth"
)

// Handler handles WebSocket connections.
type Handler struct {
	hub      *Hub
	chatRepo domain.ChatRepository
	authPr   auth.Portal
	logger   *slog.Logger
}

// NewHandler creates a new WebSocket handler.
func NewHandler(
	hub *Hub,
	chatRepo domain.ChatRepository,
	authPr auth.Portal,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		hub:      hub,
		chatRepo: chatRepo,
		authPr:   authPr,
		logger:   logger,
	}
}

// ServeHTTP handles the WebSocket upgrade and connection.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get token from query params
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	// Validate token
	authUser, err := h.authPr.ValidateToken(r.Context(), tokenString)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// Get user's chat IDs for subscription
	chatIDs, err := h.chatRepo.GetUserChatIDs(r.Context(), authUser.ID)
	if err != nil {
		h.logger.Error("failed to get user chat IDs",
			"user_id", authUser.ID,
			"error", err,
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Accept WebSocket connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Allow connections from any origin (configure appropriately for production)
	})
	if err != nil {
		h.logger.Error("failed to accept websocket connection",
			"user_id", authUser.ID,
			"error", err,
		)
		return
	}

	h.logger.Info("websocket connection established",
		"user_id", authUser.ID,
		"chat_count", len(chatIDs),
	)

	// Create client
	client := NewClient(h.hub, conn, authUser.ID, chatIDs, h.logger)

	// Register client with hub
	h.hub.Register(client)

	// Broadcast online status
	h.broadcastPresence(authUser.ID, true)

	// Run client (blocks until connection closes)
	client.Run(r.Context())

	// Broadcast offline status after connection closes
	h.broadcastPresence(authUser.ID, false)

	h.logger.Info("websocket connection closed",
		"user_id", authUser.ID,
	)
}

// broadcastPresence broadcasts user online/offline status to their contacts.
func (h *Handler) broadcastPresence(userID int, online bool) {
	eventType := EventPresenceOffline
	if online {
		eventType = EventPresenceOnline
	}

	// Get user's chats and broadcast to all participants
	chatIDs, err := h.chatRepo.GetUserChatIDs(context.Background(), userID)
	if err != nil {
		h.logger.Error("failed to get user chat IDs for presence broadcast",
			"user_id", userID,
			"error", err,
		)
		return
	}

	event := &Event{
		Type: eventType,
		Payload: PresencePayload{
			UserID: userID,
			Online: online,
		},
	}

	// Broadcast to all chats the user is part of
	for _, chatID := range chatIDs {
		h.hub.BroadcastToChat(chatID, event, userID)
	}
}
