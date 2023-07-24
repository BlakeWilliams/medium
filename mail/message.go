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
	SentAt      time.Time
	mailer      *Mailer
	Body        string
	ContentType string
	Contents    []MessageBody
}

type MessageBody struct {
	ContentType string
	Body        string
}

func (m *Message) Template(template string, data map[string]any) error {
	var b bytes.Buffer
	err := m.mailer.renderer.Render(&b, template, data)
	if err != nil {
		return err
	}

	contentType, err := contentTypeForTemplate(template)
	if err != nil {
		return err
	}

	body := MessageBody{
		ContentType: contentType,
		Body:        b.String(),
	}

	m.Contents = append(m.Contents, body)

	err = m.generateBody()
	if err != nil {
		return err
	}

	return nil
}

func (m *Message) Multipart(templates []string, data map[string]any) error {
	for _, template := range templates {
		err := m.Template(template, data)
		if err != nil {
			return err
		}
	}

	return nil
}

// Sends an email using the passed in deliverer.
func (m *Message) Send(ctx context.Context, deliverer Deliverer) error {
	if len(m.Contents) == 0 {
		return fmt.Errorf("no message body provided. call template or multipart")
	}

	err := deliverer.SendMail(ctx, m)

	if err != nil {
		return fmt.Errorf("failed to send email: %s", err)
	}

	return nil
}

func (m *Message) generateBody() error {
	if len(m.Contents) == 0 {
		return fmt.Errorf("no content provided")
	}

	if len(m.Contents) == 1 {
		m.ContentType = m.Contents[0].ContentType
		m.Body = m.Contents[0].Body
		return nil
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	defer w.Close()

	for _, content := range m.Contents {
		header := textproto.MIMEHeader{
			"Content-Type": []string{content.ContentType},
		}
		field, err := w.CreatePart(header)
		if err != nil {
			return err
		}

		_, err = field.Write([]byte(content.Body))
		if err != nil {
			return err
		}
	}

	w.Close()

	m.ContentType = fmt.Sprintf("multipart/alternative; boundary=\"%s\"; charset=UTF-8", w.Boundary())
	m.Body = b.String()

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
