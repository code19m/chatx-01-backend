package messageuc

import (
	"chatx-01/pkg/errjon"
	"context"
)

type UseCase interface {
	GetMessagesList(ctx context.Context, req GetMessagesListReq) (*GetMessagesListResp, error)
	SendMessage(ctx context.Context, req SendMessageReq) (*SendMessageResp, error)
	EditMessage(ctx context.Context, req EditMessageReq) error
	DeleteMessage(ctx context.Context, req DeleteMessageReq) error
}

type GetMessagesListReq struct {
	ChatID int `path:"chatId"`
	Page   int `query:"page"`
	Limit  int `query:"limit"`
}

func (req GetMessagesListReq) Validate() error {
	var verr error

	if req.ChatID <= 0 {
		verr = errjon.AddFieldError(verr, "chat_id", "invalid chat id")
	}
	if req.Page < 0 {
		verr = errjon.AddFieldError(verr, "page", "page must be non-negative")
	}
	if req.Limit <= 0 || req.Limit > 100 {
		verr = errjon.AddFieldError(verr, "limit", "limit must be between 1 and 100")
	}

	return verr
}

type GetMessagesListResp struct {
	Messages []MessageDTO `json:"messages"`
	Total    int          `json:"total"`
	Page     int          `json:"page"`
	Limit    int          `json:"limit"`
}

type MessageDTO struct {
	MessageID   int     `json:"message_id"`
	ChatID      int     `json:"chat_id"`
	SenderID    int     `json:"sender_id"`
	SenderName  string  `json:"sender_name"`
	SenderImage string  `json:"sender_image,omitempty"`
	Content     string  `json:"content"`
	SentAt      string  `json:"sent_at"`
	EditedAt    *string `json:"edited_at,omitempty"`
}

type SendMessageReq struct {
	ChatID  int    `json:"chat_id"`
	Content string `json:"content"`
}

func (req SendMessageReq) Validate() error {
	var verr error

	if req.ChatID <= 0 {
		verr = errjon.AddFieldError(verr, "chat_id", "invalid chat id")
	}
	if req.Content == "" {
		verr = errjon.AddFieldError(verr, "content", "message content is required")
	}
	if len(req.Content) > 5000 {
		verr = errjon.AddFieldError(verr, "content", "message content must be 5000 characters or less")
	}

	return verr
}

type SendMessageResp struct {
	MessageID int    `json:"message_id"`
	SentAt    string `json:"sent_at"`
}

type EditMessageReq struct {
	MessageID int    `path:"messageId"`
	Content   string `json:"content"`
}

func (req EditMessageReq) Validate() error {
	var verr error

	if req.MessageID <= 0 {
		verr = errjon.AddFieldError(verr, "message_id", "invalid message id")
	}
	if req.Content == "" {
		verr = errjon.AddFieldError(verr, "content", "message content is required")
	}
	if len(req.Content) > 5000 {
		verr = errjon.AddFieldError(verr, "content", "message content must be 5000 characters or less")
	}

	return verr
}

type DeleteMessageReq struct {
	MessageID int `path:"messageId"`
}

func (req DeleteMessageReq) Validate() error {
	var verr error

	if req.MessageID <= 0 {
		verr = errjon.AddFieldError(verr, "message_id", "invalid message id")
	}

	return verr
}
