package notificationuc

import (
	"chatx-01/pkg/errs"
	"context"
)

type UseCase interface {
	GetUnreadMessagesCount(
		ctx context.Context,
		req GetUnreadMessagesCountReq,
	) (*GetUnreadMessagesCountResp, error)
	GetUnreadMessagesCountByChat(
		ctx context.Context,
		req GetUnreadMessagesCountByChatReq,
	) (*GetUnreadMessagesCountByChatResp, error)
	MarkMessagesAsRead(ctx context.Context, req MarkMessagesAsReadReq) error
	GetOnlineStatusByUsers(
		ctx context.Context,
		req GetOnlineStatusByUsersReq,
	) (*GetOnlineStatusByUsersResp, error)
}

// GetUnreadMessagesCount request/response.
type GetUnreadMessagesCountReq struct {
}

func (req GetUnreadMessagesCountReq) Validate() error {
	return nil
}

type GetUnreadMessagesCountResp struct {
	TotalUnreadCount int `json:"total_unread_count"`
}

// GetUnreadMessagesCountByChat request/response.
type GetUnreadMessagesCountByChatReq struct {
	ChatID int `path:"chatId"`
}

func (req GetUnreadMessagesCountByChatReq) Validate() error {
	var verr error

	if req.ChatID <= 0 {
		verr = errs.AddFieldError(verr, "chat_id", "invalid chat id")
	}

	return verr
}

type GetUnreadMessagesCountByChatResp struct {
	ChatID      int `json:"chat_id"`
	UnreadCount int `json:"unread_count"`
}

// MarkMessagesAsRead request.
type MarkMessagesAsReadReq struct {
	ChatID    int `json:"chat_id"`
	MessageID int `json:"message_id"`
}

func (req MarkMessagesAsReadReq) Validate() error {
	var verr error

	if req.ChatID <= 0 {
		verr = errs.AddFieldError(verr, "chat_id", "invalid chat id")
	}
	if req.MessageID <= 0 {
		verr = errs.AddFieldError(verr, "message_id", "invalid message id")
	}

	return verr
}

// GetOnlineStatusByUsers request/response.
type GetOnlineStatusByUsersReq struct {
	UserIDs []int `json:"user_ids"`
}

func (req GetOnlineStatusByUsersReq) Validate() error {
	var verr error

	if len(req.UserIDs) == 0 {
		verr = errs.AddFieldError(verr, "user_ids", "at least one user id is required")
	}
	if len(req.UserIDs) > 100 {
		verr = errs.AddFieldError(verr, "user_ids", "cannot check more than 100 users at once")
	}

	return verr
}

type GetOnlineStatusByUsersResp struct {
	Statuses []UserOnlineStatus `json:"statuses"`
}

type UserOnlineStatus struct {
	UserID   int     `json:"user_id"`
	IsOnline bool    `json:"is_online"`
	LastSeen *string `json:"last_seen,omitempty"` // Only present if user is offline
}
