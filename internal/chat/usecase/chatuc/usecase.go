package chatuc

import (
	authdomain "chatx-01-backend/internal/auth/domain"
	"chatx-01-backend/internal/chat/domain"
	"chatx-01-backend/internal/portal/auth"
	"chatx-01-backend/pkg/errs"
	"context"
	"errors"
	"time"
)

type useCase struct {
	chatRepo    domain.ChatRepository
	messageRepo domain.MessageRepository
	userRepo    authdomain.UserRepository
	authPr      auth.Auth
}

// New creates a new chat use case.
func New(
	chatRepo domain.ChatRepository,
	messageRepo domain.MessageRepository,
	userRepo authdomain.UserRepository,
	authPr auth.Auth,
) UseCase {
	return &useCase{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
		authPr:      authPr,
	}
}

func (uc *useCase) GetDMsList(ctx context.Context, req GetDMsListReq) (*GetDMsListResp, error) {
	const op = "chatuc.GetDMsList"

	authUser, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	userID := authUser.ID

	offset := req.Page * req.Limit
	chats, total, err := uc.chatRepo.GetDMsListByUser(ctx, userID, offset, req.Limit)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	// TODO: Repository should return enriched data with user info, last message, unread count
	// For now returning placeholder structure
	dmItems := make([]DMListItem, len(chats))
	for i, chat := range chats {
		dmItems[i] = DMListItem{
			ChatID: chat.ID,
			// OtherUserID, OtherUsername, LastMessage etc. need to be populated
			// This requires complex JOIN queries in repository
		}
	}

	return &GetDMsListResp{
		DMs:   dmItems,
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

func (uc *useCase) GetGroupsList(ctx context.Context, req GetGroupsListReq) (*GetGroupsListResp, error) {
	const op = "chatuc.GetGroupsList"

	authUser, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	userID := authUser.ID

	offset := req.Page * req.Limit
	chats, total, err := uc.chatRepo.GetGroupsListByUser(ctx, userID, offset, req.Limit)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	// TODO: Repository should return enriched data with participant count, last message, unread count
	groupItems := make([]GroupListItem, len(chats))
	for i, chat := range chats {
		groupItems[i] = GroupListItem{
			ChatID:    chat.ID,
			Name:      chat.Name,
			CreatorID: chat.CreatorID,
			// ParticipantCount, LastMessage etc. need to be populated
			// This requires complex JOIN queries in repository
		}
	}

	return &GetGroupsListResp{
		Groups: groupItems,
		Total:  total,
		Page:   req.Page,
		Limit:  req.Limit,
	}, nil
}

func (uc *useCase) GetChat(ctx context.Context, req GetChatReq) (*GetChatResp, error) {
	const op = "chatuc.GetChat"

	authUser, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	userID := authUser.ID

	// Check if chat exists and user is participant
	chat, err := uc.chatRepo.GetByID(ctx, req.ChatID)
	if err != nil {
		return nil, errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("chat_id", "chat not found"))
	}

	isParticipant, err := uc.chatRepo.IsParticipant(ctx, req.ChatID, userID)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	if !isParticipant {
		return nil, errs.Wrap(op, domain.ErrNotParticipant)
	}

	// Get participants
	participants, err := uc.chatRepo.GetParticipants(ctx, req.ChatID)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	// Enrich with user data
	participantDTOs := make([]ChatParticipantDTO, len(participants))
	for i, p := range participants {
		user, err := uc.userRepo.GetByID(ctx, p.UserID)
		if err != nil {
			return nil, errs.Wrap(op, err)
		}

		participantDTOs[i] = ChatParticipantDTO{
			UserID:    user.ID,
			Username:  user.Username,
			ImagePath: user.ImagePath,
			JoinedAt:  p.JoinedAt.Format(time.RFC3339),
		}
	}

	return &GetChatResp{
		ChatID:       chat.ID,
		Type:         string(chat.Type),
		Name:         chat.Name,
		CreatorID:    chat.CreatorID,
		Participants: participantDTOs,
		CreatedAt:    chat.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (uc *useCase) CreateDM(ctx context.Context, req CreateDMReq) (*CreateDMResp, error) {
	const op = "chatuc.CreateDM"

	authUser, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	userID := authUser.ID

	// Check if trying to create DM with self
	if userID == req.OtherUserID {
		return nil, errs.Wrap(op, domain.ErrCannotMessageSelf)
	}

	// Check if other user exists
	_, err = uc.userRepo.GetByID(ctx, req.OtherUserID)
	if err != nil {
		return nil, errs.ReplaceOn(
			err,
			errs.ErrNotFound,
			errs.NewNotFoundError("other_user_id", "user not found"),
		)
	}

	// Check if DM already exists
	existingChat, err := uc.chatRepo.GetDMByParticipants(ctx, userID, req.OtherUserID)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return nil, errs.Wrap(op, err)
	}
	if existingChat != nil {
		return nil, errs.Wrap(op, domain.ErrDMAlreadyExists)
	}

	// Create chat
	chat := &domain.Chat{
		Type:      domain.ChatTypeDirect,
		CreatorID: userID,
		CreatedAt: time.Now(),
	}

	if err := uc.chatRepo.Create(ctx, chat); err != nil {
		return nil, errs.Wrap(op, err)
	}

	// Add both participants
	now := time.Now()
	if err := uc.chatRepo.AddParticipant(ctx, &domain.ChatParticipant{
		ChatID:   chat.ID,
		UserID:   userID,
		JoinedAt: now,
	}); err != nil {
		return nil, errs.Wrap(op, err)
	}

	if err := uc.chatRepo.AddParticipant(ctx, &domain.ChatParticipant{
		ChatID:   chat.ID,
		UserID:   req.OtherUserID,
		JoinedAt: now,
	}); err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &CreateDMResp{
		ChatID: chat.ID,
	}, nil
}

func (uc *useCase) CreateGroup(ctx context.Context, req CreateGroupReq) (*CreateGroupResp, error) {
	const op = "chatuc.CreateGroup"

	authUser, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}
	userID := authUser.ID

	// Validate all participant IDs exist
	for _, participantID := range req.ParticipantIDs {
		_, err := uc.userRepo.GetByID(ctx, participantID)
		if err != nil {
			return nil, errs.ReplaceOn(
				err,
				errs.ErrNotFound,
				errs.NewNotFoundError("participant_ids", "one or more participants not found"),
			)
		}
	}

	// Create chat
	chat := &domain.Chat{
		Type:      domain.ChatTypeGroup,
		Name:      req.Name,
		CreatorID: userID,
		CreatedAt: time.Now(),
	}

	if err := uc.chatRepo.Create(ctx, chat); err != nil {
		return nil, errs.Wrap(op, err)
	}

	// Add creator as participant
	now := time.Now()
	if err := uc.chatRepo.AddParticipant(ctx, &domain.ChatParticipant{
		ChatID:   chat.ID,
		UserID:   userID,
		JoinedAt: now,
	}); err != nil {
		return nil, errs.Wrap(op, err)
	}

	// Add other participants
	for _, participantID := range req.ParticipantIDs {
		if participantID == userID {
			continue // Skip creator, already added
		}

		if err := uc.chatRepo.AddParticipant(ctx, &domain.ChatParticipant{
			ChatID:   chat.ID,
			UserID:   participantID,
			JoinedAt: now,
		}); err != nil {
			return nil, errs.Wrap(op, err)
		}
	}

	return &CreateGroupResp{
		ChatID: chat.ID,
	}, nil
}
