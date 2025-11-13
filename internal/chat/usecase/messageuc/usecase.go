package messageuc

import (
	authdomain "chatx-01/internal/auth/domain"
	"chatx-01/internal/chat/domain"
	"chatx-01/internal/portal/auth"
	"chatx-01/pkg/errs"
	"context"
	"time"
)

type useCase struct {
	chatRepo    domain.ChatRepository
	messageRepo domain.MessageRepository
	userRepo    authdomain.UserRepository
	authPr      auth.Auth
}

// New creates a new message use case.
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

func (uc *useCase) GetMessagesList(ctx context.Context, req GetMessagesListReq) (*GetMessagesListResp, error) {
	const op = "messageuc.GetMessagesList"

	authUser, err := uc.authPr.GetAuthUser(ctx)
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

	offset := req.Page * req.Limit
	messages, total, err := uc.messageRepo.List(ctx, req.ChatID, offset, req.Limit)
	if err != nil {
		return nil, errs.Wrap(op, err)
	}

	// Enrich messages with sender data
	messageDTOs := make([]MessageDTO, len(messages))
	for i, msg := range messages {
		user, err := uc.userRepo.GetByID(ctx, msg.SenderID)
		if err != nil {
			return nil, errs.Wrap(op, err)
		}

		var editedAt *string
		if msg.EditedAt != nil {
			formatted := msg.EditedAt.Format(time.RFC3339)
			editedAt = &formatted
		}

		messageDTOs[i] = MessageDTO{
			MessageID:   msg.ID,
			ChatID:      msg.ChatID,
			SenderID:    msg.SenderID,
			SenderName:  user.Username,
			SenderImage: user.ImagePath,
			Content:     msg.Content,
			SentAt:      msg.SentAt.Format(time.RFC3339),
			EditedAt:    editedAt,
		}
	}

	return &GetMessagesListResp{
		Messages: messageDTOs,
		Total:    total,
		Page:     req.Page,
		Limit:    req.Limit,
	}, nil
}

func (uc *useCase) SendMessage(ctx context.Context, req SendMessageReq) (*SendMessageResp, error) {
	const op = "messageuc.SendMessage"

	authUser, err := uc.authPr.GetAuthUser(ctx)
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

	// Create message
	message := &domain.Message{
		ChatID:   req.ChatID,
		SenderID: userID,
		Content:  req.Content,
		SentAt:   time.Now(),
	}

	if err := uc.messageRepo.Create(ctx, message); err != nil {
		return nil, errs.Wrap(op, err)
	}

	return &SendMessageResp{
		MessageID: message.ID,
		SentAt:    message.SentAt.Format(time.RFC3339),
	}, nil
}

func (uc *useCase) EditMessage(ctx context.Context, req EditMessageReq) error {
	const op = "messageuc.EditMessage"

	authUser, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return errs.Wrap(op, err)
	}
	userID := authUser.ID

	// Get message
	message, err := uc.messageRepo.GetByID(ctx, req.MessageID)
	if err != nil {
		return errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("message_id", "message not found"))
	}

	// Check if user is the sender
	if message.SenderID != userID {
		return errs.Wrap(op, domain.ErrNotMessageOwner)
	}

	// Update message
	message.Content = req.Content
	now := time.Now()
	message.EditedAt = &now

	if err := uc.messageRepo.Update(ctx, message); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}

func (uc *useCase) DeleteMessage(ctx context.Context, req DeleteMessageReq) error {
	const op = "messageuc.DeleteMessage"

	authUser, err := uc.authPr.GetAuthUser(ctx)
	if err != nil {
		return errs.Wrap(op, err)
	}
	userID := authUser.ID

	// Get message
	message, err := uc.messageRepo.GetByID(ctx, req.MessageID)
	if err != nil {
		return errs.ReplaceOn(err, errs.ErrNotFound, errs.NewNotFoundError("message_id", "message not found"))
	}

	// Check if user is the sender
	if message.SenderID != userID {
		return errs.Wrap(op, domain.ErrNotMessageOwner)
	}

	if err := uc.messageRepo.Delete(ctx, req.MessageID); err != nil {
		return errs.Wrap(op, err)
	}

	return nil
}
