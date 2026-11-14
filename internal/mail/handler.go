package mail

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/longdnh/onward-notification/internal/event"
)

type Handler struct {
	composer Composer
	sender   *Sender
}

func NewHandler(c Composer, s *Sender) *Handler {
	return &Handler{composer: c, sender: s}
}

func (h *Handler) Handle(ctx context.Context, m event.TokenMail) error {
	now := time.Now()

	if !m.ExpiresAt.After(now) {
		log.Printf("skip %s: expired at %s", m.MessageID, m.ExpiresAt)
		return nil
	}

	var msg Message
	switch m.Type {
	case event.TypeEmailVerification:
		msg = h.composer.Verification(m.Email, m.Token, m.ExpiresAt, now)
	case event.TypePasswordReset:
		msg = h.composer.PasswordReset(m.Email, m.Token, m.ExpiresAt, now)
	default:
		return fmt.Errorf("unknown type %q", m.Type)
	}

	return h.sender.Send(ctx, msg)
}
