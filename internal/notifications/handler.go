package notifications

import (
	"chatx-01-backend/internal/events"
	"chatx-01-backend/internal/notifications/usecase"
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

// Handler handles notification events.
type Handler struct {
	notificationUC usecase.UseCase
}

// NewHandler creates a new notification handler.
func NewHandler(notificationUC usecase.UseCase) *Handler {
	return &Handler{
		notificationUC: notificationUC,
	}
}

// HandleUserRegistration handles user registration events.
func (h *Handler) HandleUserRegistration(ctx context.Context, msg *sarama.ConsumerMessage) error {
	// Parse event
	event, err := events.UnmarshalUserRegisteredEvent(msg.Value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Forward to use case
	return h.notificationUC.SendWelcomeEmail(ctx, usecase.SendWelcomeEmailReq{
		Email:    event.Email,
		Username: event.Username,
		Password: event.Password,
	})
}
