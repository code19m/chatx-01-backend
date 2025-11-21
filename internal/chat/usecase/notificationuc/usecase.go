package notificationuc

import (
	"chatx-01-backend/internal/chat/controller/ws"
	"chatx-01-backend/internal/chat/domain"
	"chatx-01-backend/internal/portal/auth"
	"chatx-01-backend/pkg/errs"
	"context"
	"time"
)

// OnlineChecker provides online status checking capability.
type OnlineChecker interface {
	IsUserOnline(userID int) bool
	GetOnlineUsers(userIDs []int) []int
}

type useCase struct {
	chatRepo      domain.ChatRepository
	messageRepo   domain.MessageRepository
	authPortal    auth.Portal
	broadcaster   ws.Broadcaster
	onlineChecker OnlineChecker
}

// New creates a new notification use case.
func New(
	chatRepo domain.ChatRepository,
	messageRepo domain.MessageRepository,
	authPortal auth.Portal,
	broadcaster ws.Broadcaster,
	onlineChecker OnlineChecker,
) UseCase {
	return &useCase{
		chatRepo:      chatRepo,
		messageRepo:   messageRepo,
		authPortal:    authPortal,
		broadcaster:   broadcaster,
		onlineChecker: onlineChecker,
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

	// Broadcast read receipt via WebSocket
	uc.broadcaster.BroadcastReadReceipt(req.ChatID, userID, req.MessageID, time.Now())

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

	// Get online users from WebSocket hub
	onlineUserIDs := uc.onlineChecker.GetOnlineUsers(req.UserIDs)
	onlineSet := make(map[int]bool, len(onlineUserIDs))
	for _, id := range onlineUserIDs {
		onlineSet[id] = true
	}

	statuses := make([]UserOnlineStatus, len(req.UserIDs))
	for i, userID := range req.UserIDs {
		statuses[i] = UserOnlineStatus{
			UserID:   userID,
			IsOnline: onlineSet[userID],
			LastSeen: nil, // TODO: Implement last seen tracking with Redis
		}
	}

	return &GetOnlineStatusByUsersResp{
		Statuses: statuses,
	}, nil
}
