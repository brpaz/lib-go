// Package mailer provides a [Mailer] abstraction for sending email, plus
// implementations for common scenarios.
//
// # Sending mail
//
// Build a [Message] and pass it to Send:
//
//	err := m.Send(ctx, mailer.Message{
//	    From:    "noreply@example.com",
//	    To:      []string{user.Email},
//	    Subject: "Welcome!",
//	    Body:    "<p>Thanks for signing up.</p>",
//	    HTML:    true,
//	})
//
// # Implementations
//
// [SMTP] sends mail through an SMTP server using the standard library:
//
//	auth := smtp.PlainAuth("", "user", "pass", "smtp.example.com")
//	m    := mailer.NewSMTP("smtp.example.com:587", auth)
//
// # Testing
//
// Pass [Noop] when a component needs a [Mailer] but the test doesn't care
// about email — every Send is discarded:
//
//	svc := NewSignupService(mailer.Noop{})
//
// Pass [Fake] to assert which messages a service sent, without wiring a real
// transport:
//
//	m   := mailer.NewFake()
//	svc := NewSignupService(m)
//
//	_ = svc.Register(ctx, user)
//	assert.Equal(t, []mailer.Message{
//	    {From: "noreply@example.com", To: []string{user.Email}, Subject: "Welcome!"},
//	}, m.Sent())
//
// # Swapping the implementation
//
// Code should depend on the [Mailer] interface rather than a concrete type.
// This allows swapping [SMTP] for a transactional email provider's API (e.g.
// Postmark, Resend, SES) at the wiring layer without touching the rest of the
// codebase.
package mailer
