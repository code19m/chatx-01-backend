package usecase

import "context"

type UseCase interface {
	SendWelcomeEmail(ctx context.Context, req SendWelcomeEmailReq) error
}

type SendWelcomeEmailReq struct {
	Email    string
	Username string
	Password string
}
