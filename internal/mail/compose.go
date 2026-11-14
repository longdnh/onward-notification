// Package mail composes the auth mails and sends them over SMTP.
package mail

import (
	"fmt"
	"net/url"
	"time"
)

// Message is a plain-text mail ready to be sent.
type Message struct {
	To      string
	Subject string
	Body    string
}

// Composer builds the auth mails. The token goes in the URL query only; the
// frontend page extracts it and POSTs it to Account. Mirrors AuthMailComposer
// in onward-be so the two delivery paths produce the same mail.
type Composer struct {
	frontendBaseURL string // without trailing slash
}

func NewComposer(frontendBaseURL string) Composer {
	return Composer{frontendBaseURL: frontendBaseURL}
}

func (c Composer) Verification(to, token string, expiresAt, now time.Time) Message {
	body := fmt.Sprintf(`Welcome to Onward. Confirm your email address by opening this link: %s
		The link is valid for %s. If you did not create an account, ignore this message.
		`, c.link("/verify-email", token), humanDuration(expiresAt.Sub(now)))
			return Message{To: to, Subject: "Confirm your email address", Body: body}
}

func (c Composer) PasswordReset(to, token string, expiresAt, now time.Time) Message {
	body := fmt.Sprintf(`A password reset was requested for your Onward account.

Choose a new password by opening this link:
%s

The link is valid for %s and can be used once. If you did not request this, ignore this message;
your password stays unchanged.
`, c.link("/reset-password", token), humanDuration(expiresAt.Sub(now)))
	return Message{To: to, Subject: "Reset your password", Body: body}
}

func (c Composer) link(path, token string) string {
	return c.frontendBaseURL + path + "?token=" + url.QueryEscape(token)
}

// humanDuration rounds up to whole minutes: the message is handled some time
// after the token was issued, so a 24h token would otherwise show as 23h59m.
func humanDuration(d time.Duration) string {
	minutes := max(int64(d.Seconds()+59)/60, 1)
	if minutes%60 == 0 {
		hours := minutes / 60
		return fmt.Sprintf("%d hour%s", hours, plural(hours))
	}
	return fmt.Sprintf("%d minute%s", minutes, plural(minutes))
}

func plural(n int64) string {
	if n == 1 {
		return ""
	}
	return "s"
}
