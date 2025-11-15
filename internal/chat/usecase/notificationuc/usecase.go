package notificationuc

import (
	"chatx-01-backend/internal/chat/domain"
	"chatx-01-backend/internal/portal/auth"
	"chatx-01-backend/pkg/errs"
	"context"
)

type useCase struct {
	chatRepo    domain.ChatRepository
	messageRepo domain.MessageRepository
	authPortal  auth.Portal
}

// New creates a new notification use case.
func New(
	chatRepo domain.ChatRepository,
	messageRepo domain.MessageRepository,
	authPortal auth.Portal,
) UseCase {
	return &useCase{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		authPortal:  authPortal,
	}
}

func (uc *useCase) GetUnreadMessagesCount(
	ctx context.Context,
	req GetUnreadMessagesCountReq,
) (*GetUnreadMessagesCountResp, error) {
	const op = "notificationuc.GetUnreadMessagesCount"

	authUser, err := uc.authPortal.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	userID := authUser.ID

	totalCount, err := uc.messageRepo.GetTotalUnreadCount(ctx, userID)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &GetUnreadMessagesCountResp{
		TotalUnreadCount: totalCount,
	}, nil
}

func (uc *useCase) GetUnreadMessagesCountByChat(
	ctx context.Context,
	req GetUnreadMessagesCountByChatReq,
) (*GetUnreadMessagesCountByChatResp, error) {
	const op = "notificationuc.GetUnreadMessagesCountByChat"

	authUser, err := uc.authPortal.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	userID := authUser.ID

	// Check if user is participant
	isParticipant, err := uc.chatRepo.IsParticipant(ctx, req.ChatID, userID)
	if err != nil {
		return nil, errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("chat_id", "chat not found"))
	}
	if !isParticipant {
		return nil, errs.Wrap(op, domain.ErrNotParticipant)
	}

	unreadCount, err := uc.messageRepo.GetUnreadCountByChat(ctx, req.ChatID, userID)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &GetUnreadMessagesCountByChatResp{
		ChatID:      req.ChatID,
		UnreadCount: unreadCount,
	}, nil
}

func (uc *useCase) MarkMessagesAsRead(ctx context.Context, req MarkMessagesAsReadReq) error {
	const op = "notificationuc.MarkMessagesAsRead"

	authUser, err := uc.authPortal.GetAuthUser(ctx)
	if err != nil {
		return errs.Wrap(op, err)
	}
	userID := authUser.ID

	// Check if user is participant
	isParticipant, err := uc.chatRepo.IsParticipant(ctx, req.ChatID, userID)
	if err != nil {
		return errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("chat_id", "chat not found"))
	}
	if !isParticipant {
		return errs.Wrap(op, domain.ErrNotParticipant)
	}

	// Verify message exists and belongs to chat
	message, err := uc.messageRepo.GetByID(ctx, req.MessageID)
	if err != nil {
		return errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("message_id", "message not found"))
	}

	if message.ChatID != req.ChatID {
		return errs.Wrap(op, domain.ErrMessageNotInChat)
	}

	// Update last read message
	if err := uc.chatRepo.UpdateLastRead(ctx, req.ChatID, userID, req.MessageID); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (uc *useCase) GetOnlineStatusByUsers(
	ctx context.Context,
	req GetOnlineStatusByUsersReq,
) (*GetOnlineStatusByUsersResp, error) {
	const op = "notificationuc.GetOnlineStatusByUsers"

	_, err := uc.authPortal.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	// TODO: Implement online status tracking
	// This requires a separate mechanism (Redis, WebSocket tracking, etc.)
	// For now, returning placeholder with all users offline
	statuses := make([]UserOnlineStatus, len(req.UserIDs))
	for i, userID := range req.UserIDs {
		statuses[i] = UserOnlineStatus{
			UserID:   userID,
			IsOnline: false,
			LastSeen: nil, // Will be populated when tracking is implemented
		}
	}

	return &GetOnlineStatusByUsersResp{
		Statuses: statuses,
	}, nil
}
