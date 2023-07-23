package mail

import (
	"context"
	"time"

	"github.com/blakewilliams/bat"
)

// Mailer stores state required to connect to a mail server and send emails. It
// requires a view.Renderer so that it can send HTML emails.
type Mailer struct {
	DevMode  bool
	renderer *bat.Engine
	From     string

	// SentMessages is slice of mail that is collected when DevMode is true.
	SentMail []Message

	deliverer Deliverer
}

// Represents a type that can be used to send emails to mail servers.
type Deliverer interface {
	SendMail(context.Context, *Message) error
}

// Creates a new mailer, accepting a renderer which is used to render HTML
// emails, the mailer host, and the mailer auth.
func New(deliverer Deliverer, renderer *bat.Engine) *Mailer {
	mailer := &Mailer{
		deliverer: deliverer,
		renderer:  renderer,
		From:      "noreply@noreply.net",
	}

	return mailer
}

// Creates a new message that can be modified and delivered via Send
func (m *Mailer) NewMessage(subject string, to ...string) *Message {
	return &Message{
		To:      to,
		From:    m.From,
		Subject: subject,
		mailer:  m,
	}
}

// Sends an email using the mailer's host and auth.
func (m *Mailer) Send(ctx context.Context, msg *Message) error {
	if m.DevMode {
		m.SentMail = append(m.SentMail, *msg)
	} else {
		err := msg.Send(ctx, m.deliverer)

		if err != nil {
			return err
		}
	}

	return nil
}

// Creates a new email message that can be sent.
func NewMail(to []string, subject string, from string, body string) *Message {
	return &Message{
		To:      to,
		From:    from,
		Subject: subject,
		Body:    body,
		SentAt:  time.Now(),
	}
}

func (m *Mailer) ResetSentMail() {
	m.SentMail = make([]Message, 0)
}
