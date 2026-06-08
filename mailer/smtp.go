package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTP is a [Mailer] that sends mail through an SMTP server using the
// standard library's net/smtp package.
type SMTP struct {
	addr string
	auth smtp.Auth
}

// NewSMTP returns a [SMTP] mailer that dials addr (host:port) and
// authenticates with auth. Pass nil for auth to skip authentication.
func NewSMTP(addr string, auth smtp.Auth) *SMTP {
	return &SMTP{addr: addr, auth: auth}
}

// Send implements [Mailer]. It builds a minimal MIME message from msg and
// hands it off via [smtp.SendMail].
func (m *SMTP) Send(ctx context.Context, msg Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(msg.To) == 0 {
		return ErrNoRecipients
	}

	return smtp.SendMail(m.addr, m.auth, msg.From, msg.To, buildMIME(msg))
}

func buildMIME(msg Message) []byte {
	contentType := "text/plain"
	if msg.HTML {
		contentType = "text/html"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", sanitizeHeader(msg.From))
	fmt.Fprintf(&b, "To: %s\r\n", sanitizeHeader(strings.Join(msg.To, ", ")))
	fmt.Fprintf(&b, "Subject: %s\r\n", sanitizeHeader(msg.Subject))
	fmt.Fprintf(&b, "Content-Type: %s; charset=\"UTF-8\"\r\n", contentType)
	b.WriteString("\r\n")
	b.WriteString(msg.Body)

	return []byte(b.String())
}

// sanitizeHeader strips CR/LF from a header value to prevent SMTP header
// injection (e.g. a malicious subject or address smuggling extra headers or
// recipients into the message).
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")

	return s
}
