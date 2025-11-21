package events

import (
	"encoding/json"
	"fmt"
)

// UserRegisteredEvent represents a user registration event.
type UserRegisteredEvent struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// MarshalJSON marshals the event to JSON.
func (e UserRegisteredEvent) Marshal() ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}
	return data, nil
}

// UnmarshalUserRegisteredEvent unmarshals the event from JSON.
func UnmarshalUserRegisteredEvent(data []byte) (UserRegisteredEvent, error) {
	var event UserRegisteredEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return UserRegisteredEvent{}, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return event, nil
}
