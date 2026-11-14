package mail

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Sender delivers messages through one SMTP server, unauthenticated and
// without TLS: enough for Mailpit locally. A real provider needs auth and
// STARTTLS, which is a change inside this file only.
type Sender struct {
	addr string // host:port
	from string
}

func NewSender(host string, port int, from string) *Sender {
	return &Sender{addr: net.JoinHostPort(host, fmt.Sprint(port)), from: from}
}

// Send opens a connection per message. Fine at this volume; a pool comes
// later if it ever matters. The context bounds the dial only: net/smtp has
// no context support, so the rest relies on the deadline set on the socket.
func (s *Sender) Send(ctx context.Context, m Message) error {
	dialer := net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", s.addr)
	if err != nil {
		return fmt.Errorf("dial smtp %s: %w", s.addr, err)
	}
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	client, err := smtp.NewClient(conn, s.addr)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp handshake: %w", err)
	}
	defer client.Close()

	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	if err := client.Rcpt(m.To); err != nil {
		return fmt.Errorf("smtp RCPT TO: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}
	if _, err := w.Write([]byte(s.render(m))); err != nil {
		return fmt.Errorf("smtp write body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp end of data: %w", err)
	}
	return client.Quit()
}

// render produces the raw RFC 5322 message: headers, blank line, body.
func (s *Sender) render(m Message) string {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", s.from)
	fmt.Fprintf(&b, "To: %s\r\n", m.To)
	fmt.Fprintf(&b, "Subject: %s\r\n", m.Subject)
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("\r\n")
	// SMTP wants CRLF line endings; the composer writes LF.
	b.WriteString(strings.ReplaceAll(m.Body, "\n", "\r\n"))
	return b.String()
}
