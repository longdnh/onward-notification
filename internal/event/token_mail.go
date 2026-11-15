package event

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

const (
	TypeEmailVerification = "onward.auth.email-verification-requested.v1"
	TypePasswordReset     = "onward.auth.password-reset-requested.v1"
)

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

	if m.Type != TypeEmailVerification && m.Type != TypePasswordReset {
		return fmt.Errorf("unknown type %q", m.Type)
	}

	if m.MessageID == "" {
		return errors.New("MessageId is empty")
	}
	return nil
}

func Decode(body []byte) (TokenMail, error) {
	var m TokenMail

	if err := json.Unmarshal(body, &m); err != nil {
		return TokenMail{}, fmt.Errorf("invalid json: %w", err)
	}

	if err := m.Validate(); err != nil {
		return TokenMail{}, err
	}
	return m, nil
}
