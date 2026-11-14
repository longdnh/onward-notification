package rabbit

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type TokenMail struct {
	MessageID  string    `json:"messageId"`
	Type       string    `json:"type"`
	OccurredAt time.Time `json:"occurredAt"`
	Email      string    `json:"email"`
	Token      string    `json:"token"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

func (m TokenMail) Validate() error {
	if m.Email == "" {
		return errors.New("Email is empty")
	}

	if m.Token == "" {
		return errors.New("Token is empty")
	}

	if m.ExpiresAt.IsZero() {
		return errors.New("ExpiresAt is empty")
	}
	return nil
}

func decode(body []byte) (TokenMail, error) {
	var m TokenMail

	if err := json.Unmarshal(body, &m); err != nil {
		return TokenMail{}, fmt.Errorf("invalid json: %w", err)
	}

	if err := m.Validate(); err != nil {
		return TokenMail{}, err
	}
	return m, nil
}
