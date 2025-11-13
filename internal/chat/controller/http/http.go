package http

import (
	"chatx-01-backend/internal/chat/usecase/chatuc"
	"chatx-01-backend/internal/chat/usecase/messageuc"
	"chatx-01-backend/internal/chat/usecase/notificationuc"
	"net/http"
)

type ctrl struct {
	mux    *http.ServeMux
	prefix string

	chatUsecase         chatuc.UseCase
	messageUsecase      messageuc.UseCase
	notificationUsecase notificationuc.UseCase
}

func New(
	mux *http.ServeMux,
	prefix string,
	chatUsecase chatuc.UseCase,
	messageUsecase messageuc.UseCase,
	notificationUsecase notificationuc.UseCase,
) *ctrl {
	c := &ctrl{
		mux:                 mux,
		prefix:              prefix,
		chatUsecase:         chatUsecase,
		messageUsecase:      messageUsecase,
		notificationUsecase: notificationUsecase,
	}

	c.registerHandlers()
	return c
}

// registerHandlers registers all handlers.
func (c *ctrl) registerHandlers() {
	// Chat endpoints
	c.register(http.MethodGet, "/chats/dms", http.HandlerFunc(c.getDMsList))
	c.register(http.MethodGet, "/chats/groups", http.HandlerFunc(c.getGroupsList))
	c.register(http.MethodGet, "/chats/{chatId}", http.HandlerFunc(c.getChat))
	c.register(http.MethodPost, "/chats/dms", http.HandlerFunc(c.createDM))
	c.register(http.MethodPost, "/chats/groups", http.HandlerFunc(c.createGroup))

	// Message endpoints
	c.register(http.MethodGet, "/chats/{chatId}/messages", http.HandlerFunc(c.getMessagesList))
	c.register(http.MethodPost, "/messages", http.HandlerFunc(c.sendMessage))
	c.register(http.MethodPut, "/messages/{messageId}", http.HandlerFunc(c.editMessage))
	c.register(http.MethodDelete, "/messages/{messageId}", http.HandlerFunc(c.deleteMessage))

	// Notification endpoints
	c.register(http.MethodGet, "/notifications/unread", http.HandlerFunc(c.getUnreadMessagesCount))
	c.register(http.MethodGet, "/chats/{chatId}/unread", http.HandlerFunc(c.getUnreadMessagesCountByChat))
	c.register(http.MethodPost, "/chats/read", http.HandlerFunc(c.markMessagesAsRead))
	c.register(http.MethodPost, "/users/online-status", http.HandlerFunc(c.getOnlineStatusByUsers))
}

func (c *ctrl) register(method string, path string, handler http.Handler) {
	fullPath := c.prefix + path
	c.mux.Handle(method+" "+fullPath, handler)
}
