package chatuc

import (
	"chatx-01/pkg/errjon"
	"context"
)

type UseCase interface {
	GetDMsList(ctx context.Context, req GetDMsListReq) (*GetDMsListResp, error)
	GetGroupsList(ctx context.Context, req GetGroupsListReq) (*GetGroupsListResp, error)
	GetChat(ctx context.Context, req GetChatReq) (*GetChatResp, error)
	CreateDM(ctx context.Context, req CreateDMReq) (*CreateDMResp, error)
	CreateGroup(ctx context.Context, req CreateGroupReq) (*CreateGroupResp, error)
}

// GetDMsList request/response.
type GetDMsListReq struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

func (req GetDMsListReq) Validate() error {
	var verr error

	if req.Page < 0 {
		verr = errjon.AddFieldError(verr, "page", "page must be non-negative")
	}
	if req.Limit <= 0 || req.Limit > 100 {
		verr = errjon.AddFieldError(verr, "limit", "limit must be between 1 and 100")
	}

	return verr
}

type GetDMsListResp struct {
	DMs   []DMListItem `json:"dms"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

type DMListItem struct {
	ChatID            int     `json:"chat_id"`
	OtherUserID       int     `json:"other_user_id"`
	OtherUsername     string  `json:"other_username"`
	OtherUserImage    string  `json:"other_user_image,omitempty"`
	LastMessageText   *string `json:"last_message_text,omitempty"`
	LastMessageSentAt *string `json:"last_message_sent_at,omitempty"`
	UnreadCount       int     `json:"unread_count"`
}

// GetGroupsList request/response.
type GetGroupsListReq struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

func (req GetGroupsListReq) Validate() error {
	var verr error

	if req.Page < 0 {
		verr = errjon.AddFieldError(verr, "page", "page must be non-negative")
	}
	if req.Limit <= 0 || req.Limit > 100 {
		verr = errjon.AddFieldError(verr, "limit", "limit must be between 1 and 100")
	}

	return verr
}

type GetGroupsListResp struct {
	Groups []GroupListItem `json:"groups"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

type GroupListItem struct {
	ChatID            int     `json:"chat_id"`
	Name              string  `json:"name"`
	CreatorID         int     `json:"creator_id"`
	ParticipantCount  int     `json:"participant_count"`
	LastMessageText   *string `json:"last_message_text,omitempty"`
	LastMessageSentAt *string `json:"last_message_sent_at,omitempty"`
	UnreadCount       int     `json:"unread_count"`
}

// GetChat request/response.
type GetChatReq struct {
	ChatID int `path:"chatId"`
}

func (req GetChatReq) Validate() error {
	var verr error

	if req.ChatID <= 0 {
		verr = errjon.AddFieldError(verr, "chat_id", "invalid chat id")
	}

	return verr
}

type GetChatResp struct {
	ChatID       int                  `json:"chat_id"`
	Type         string               `json:"type"` // "direct" or "group"
	Name         string               `json:"name,omitempty"`
	CreatorID    int                  `json:"creator_id,omitempty"`
	Participants []ChatParticipantDTO `json:"participants"`
	CreatedAt    string               `json:"created_at"`
}

type ChatParticipantDTO struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	ImagePath string `json:"image_path,omitempty"`
	JoinedAt  string `json:"joined_at"`
}

// CreateDM request/response.
type CreateDMReq struct {
	OtherUserID int `json:"other_user_id"`
}

func (req CreateDMReq) Validate() error {
	var verr error

	if req.OtherUserID <= 0 {
		verr = errjon.AddFieldError(verr, "other_user_id", "invalid other user id")
	}

	return verr
}

type CreateDMResp struct {
	ChatID int `json:"chat_id"`
}

// CreateGroup request/response.
type CreateGroupReq struct {
	Name           string `json:"name"`
	ParticipantIDs []int  `json:"participant_ids"`
}

func (req CreateGroupReq) Validate() error {
	var verr error

	if req.Name == "" {
		verr = errjon.AddFieldError(verr, "name", "group name is required")
	}
	if len(req.Name) > 100 {
		verr = errjon.AddFieldError(verr, "name", "group name must be 100 characters or less")
	}
	if len(req.ParticipantIDs) == 0 {
		verr = errjon.AddFieldError(verr, "participant_ids", "at least one participant is required")
	}

	return verr
}

type CreateGroupResp struct {
	ChatID int `json:"chat_id"`
}
