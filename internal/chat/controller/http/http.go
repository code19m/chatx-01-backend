package http

import (
	"chatx-01-backend/internal/chat/usecase/chatuc"
	"chatx-01-backend/internal/chat/usecase/messageuc"
	"chatx-01-backend/internal/chat/usecase/notificationuc"
	"chatx-01-backend/internal/portal/auth"
	"net/http"
)

type ctrl struct {
	mux    *http.ServeMux
	prefix string

	chatUsecase         chatuc.UseCase
	messageUsecase      messageuc.UseCase
	notificationUsecase notificationuc.UseCase

	authPr auth.Portal
}

func Register(
	mux *http.ServeMux,
	prefix string,
	chatUsecase chatuc.UseCase,
	messageUsecase messageuc.UseCase,
	notificationUsecase notificationuc.UseCase,
	authPr auth.Portal,
) {
	c := &ctrl{
		mux:                 mux,
		prefix:              prefix,
		chatUsecase:         chatUsecase,
		messageUsecase:      messageUsecase,
		notificationUsecase: notificationUsecase,
		authPr:              authPr,
	}

	c.registerHandlers()
}

// registerHandlers registers all handlers.
func (c *ctrl) registerHandlers() {
	// Chat endpoints
	c.register(http.MethodGet, "/chats/dms", http.HandlerFunc(c.getDMsList), c.authPr.RequireAuth())
	c.register(http.MethodGet, "/chats/groups", http.HandlerFunc(c.getGroupsList), c.authPr.RequireAuth())
	c.register(http.MethodGet, "/chats/{chat_id}", http.HandlerFunc(c.getChat), c.authPr.RequireAuth())
	c.register(http.MethodGet, "/chats/dms/check", http.HandlerFunc(c.checkDMExists), c.authPr.RequireAuth())
	c.register(http.MethodPost, "/chats/dms", http.HandlerFunc(c.createDM), c.authPr.RequireAuth())
	c.register(http.MethodPost, "/chats/groups", http.HandlerFunc(c.createGroup), c.authPr.RequireAuth())

	// Message endpoints
	c.register(http.MethodGet, "/chats/{chat_id}/messages", http.HandlerFunc(c.getMessagesList), c.authPr.RequireAuth())
	c.register(http.MethodPost, "/messages", http.HandlerFunc(c.sendMessage), c.authPr.RequireAuth())
	c.register(http.MethodPut, "/messages/{message_id}", http.HandlerFunc(c.editMessage), c.authPr.RequireAuth())
	c.register(http.MethodDelete, "/messages/{message_id}", http.HandlerFunc(c.deleteMessage), c.authPr.RequireAuth())

	// Notification endpoints
	c.register(
		http.MethodGet,
		"/notifications/unread",
		http.HandlerFunc(c.getUnreadMessagesCount),
		c.authPr.RequireAuth(),
	)
	c.register(
		http.MethodGet,
		"/chats/{chat_id}/unread",
		http.HandlerFunc(c.getUnreadMessagesCountByChat),
		c.authPr.RequireAuth(),
	)
	c.register(http.MethodPost, "/chats/read", http.HandlerFunc(c.markMessagesAsRead), c.authPr.RequireAuth())
	c.register(
		http.MethodPost,
		"/users/online-status",
		http.HandlerFunc(c.getOnlineStatusByUsers),
		c.authPr.RequireAuth(),
	)
}

func (c *ctrl) register(
	method string,
	path string,
	handler http.Handler,
	middlewares ...func(http.Handler) http.Handler,
) {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	fullPath := c.prefix + path
	c.mux.Handle(method+" "+fullPath, handler)
}
