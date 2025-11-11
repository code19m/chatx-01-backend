package domain

import "errors"

// Domain-specific errors for chat module.
var (
	ErrNotParticipant    = errors.New("user is not a participant of this chat")
	ErrDMAlreadyExists   = errors.New("direct message chat already exists")
	ErrCannotMessageSelf = errors.New("cannot create DM with yourself")
	ErrNotMessageOwner   = errors.New("user is not the owner of this message")
	ErrMessageNotInChat  = errors.New("message does not belong to this chat")
)
