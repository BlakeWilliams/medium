package mail

import (
	"bytes"
	"fmt"
	"time"

	"github.com/blakewilliams/medium/pkg/template"
)

// Mailer stores state required to connect to a mail server and send emails. It
// requires a template.Renderer so that it can send HTML emails.
type Mailer struct {
	DevMode  bool
	renderer *template.Renderer
	From     string

	// SentMessages is slice of mail that is collected when DevMode is true.
	SentMail []Mail

	deliverer Deliverer
}

// Mail is used to store the data required to send an email.
type Mail struct {
	To      []string
	From    string
	Subject string
	Body    string
	SentAt  time.Time
}

// Represents a type that can be used to send emails to mail servers.
type Deliverer interface {
	SendMail(from string, to []string, body []byte) error
}

// Creates a new mailer, accepting a renderer which is used to render HTML
// emails, the mailer host, and the mailer auth.
func New(deliverer Deliverer, renderer *template.Renderer) *Mailer {
	mailer := &Mailer{
		deliverer: deliverer,
		renderer:  renderer,
		From:      "noreply@noreply.net",
	}

	return mailer
}

// Sends an email using the mailer's host and auth.
func (m *Mailer) Send(templateName string, to string, subject string, data map[string]interface{}) error {
	buf := new(bytes.Buffer)
	err := m.renderer.Render(buf, templateName, data)
	if err != nil {
		return fmt.Errorf("failed to render template: %s", err)
	}

	mail := NewMail([]string{to}, subject, m.From, buf.String())

	if m.DevMode {
		m.SentMail = append(m.SentMail, *mail)
	} else {
		err := mail.Send(m.deliverer)

		if err != nil {
			return err
		}
	}

	return nil
}

// Creates a new email message that can be sent.
func NewMail(to []string, subject string, from string, body string) *Mail {
	return &Mail{
		To:      to,
		From:    from,
		Subject: subject,
		Body:    body,
		SentAt:  time.Now(),
	}
}

func (m *Mail) Send(deliverer Deliverer) error {
	err := deliverer.SendMail(m.From, m.To, []byte(m.Body))

	if err != nil {
		return fmt.Errorf("failed to send email: %s", err)
	}

	return nil
}
