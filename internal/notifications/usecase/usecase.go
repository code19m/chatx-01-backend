package usecase

import (
	"chatx-01-backend/pkg/email"
	"chatx-01-backend/pkg/errs"
	"context"
	"log/slog"
)

type useCase struct {
	emailSender email.Sender
}

func New(emailSender email.Sender) UseCase {
	return &useCase{
		emailSender: emailSender,
	}
}

func (uc *useCase) SendWelcomeEmail(ctx context.Context, req SendWelcomeEmailReq) error {
	const op = "notificationuc.SendWelcomeEmail"

	slog.Info("processing user registration event",
		"email", req.Email,
		"username", req.Username,
	)

	// Build welcome email
	welcomeEmail, err := email.BuildWelcomeEmail(req.Email, req.Username, req.Password)
	if err != nil {
		return errs.Wrap(op, err)
	}

	// Send email
	err = uc.emailSender.Send(welcomeEmail)
	if err != nil {
		return errs.Wrap(op, err)
	}

	slog.Info("welcome email sent successfully",
		"email", req.Email,
		"username", req.Username,
	)

	return nil
}
