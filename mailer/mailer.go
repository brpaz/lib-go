package mailer

import (
	"context"
	"errors"
)

// ErrNoRecipients is returned by Send when the message has no recipients.
var ErrNoRecipients = errors.New("mailer: message has no recipients")

// Message is an email to be sent.
type Message struct {
	// From is the sender address.
	From string

	// To lists the recipient addresses. At least one is required.
	To []string

	// Subject is the email subject line.
	Subject string

	// Body is the email content. Its format is determined by HTML.
	Body string

	// HTML indicates whether Body is HTML (true) or plain text (false).
	HTML bool
}

// Mailer sends email messages.
//
// Implementations can talk to an SMTP server (like [SMTP]) or a transactional
// email provider's API (e.g. Postmark, Resend, SES) by providing a different
// implementation at the wiring layer.
type Mailer interface {
	// Send delivers msg. It returns an error if the message could not be
	// handed off to the underlying transport.
	Send(ctx context.Context, msg Message) error
}
