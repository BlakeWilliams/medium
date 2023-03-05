package mail

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"path"
	"time"
)

// Message is used to store the data required to send an email.
type Message struct {
	To          []string
	From        string
	Subject     string
	Body        string
	SentAt      time.Time
	ContentType string
	mailer      *Mailer
}

func (m *Message) Template(template string, data map[string]any) error {
	var b bytes.Buffer
	err := m.mailer.renderer.Render(&b, template, data)
	if err != nil {
		return err
	}

	m.Body = b.String()
	contentType, err := contentTypeForTemplate(template)
	if err != nil {
		return err
	}

	m.ContentType = contentType

	return nil
}

func (m *Message) Multipart(templates []string, data map[string]any) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	defer w.Close()

	for _, template := range templates {
		var tb bytes.Buffer
		err := m.mailer.renderer.Render(&tb, template, data)
		if err != nil {
			return err
		}

		contentType, err := contentTypeForTemplate(template)
		if err != nil {
			return err
		}

		header := textproto.MIMEHeader{
			"Content-Type": []string{contentType},
		}
		field, err := w.CreatePart(header)
		if err != nil {
			return err
		}

		_, err = field.Write(tb.Bytes())

		if err != nil {
			return err
		}
	}

	w.Close()
	m.Body = b.String()

	if len(templates) > 1 {
		m.ContentType = fmt.Sprintf("multipart/mixed; boundary=\"%s\"; charset=UTF-8", w.Boundary())
		return nil
	}

	ext := path.Ext(templates[0])
	if ext == ".txt" || ext == ".text" {
		m.ContentType = "text/plain"
		return nil
	}

	m.ContentType = mime.TypeByExtension(ext)
	if m.ContentType == "" {
		return fmt.Errorf("could not determine mimetype for template %s", templates[0])
	}

	return nil
}

// Sends an email using the passed in deliverer.
func (m *Message) Send(ctx context.Context, deliverer Deliverer) error {
	err := deliverer.SendMail(ctx, m)

	if err != nil {
		return fmt.Errorf("failed to send email: %s", err)
	}

	return nil
}

func contentTypeForTemplate(name string) (string, error) {
	ext := path.Ext(name)

	if ext == ".txt" || ext == ".text" {
		return "text/plain", nil
	}

	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "", fmt.Errorf("could not determine mimetype for template %s", name)
	}

	return mimeType, nil
}
