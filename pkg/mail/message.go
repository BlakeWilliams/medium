package mail

import (
	"bytes"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"path"
	"strings"
	"time"
)

// Message is used to store the data required to send an email.
type Message struct {
	To          []string
	From        string
	Subject     string
	Body        string
	SentAt      time.Time
	contentType string
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

	m.contentType = contentType

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
		m.contentType = fmt.Sprintf("multipart/mixed; boundary=%s", w.Boundary())
		return nil
	}

	ext := path.Ext(templates[0])
	if ext == ".txt" || ext == ".text" {
		m.contentType = "text/plain"
		return nil
	}

	m.contentType = mime.TypeByExtension(ext)
	if m.contentType == "" {
		return fmt.Errorf("could not determine mimetype for template %s", templates[0])
	}

	return nil
}

// Sends an email using the passed in deliverer.
func (m *Message) Send(deliverer Deliverer) error {
	msg := new(bytes.Buffer)
	msg.WriteString("From: " + m.From + "\r\n")
	msg.WriteString("To: " + strings.Join(m.To, ",") + "\r\n")
	msg.WriteString("Subject: " + m.Subject + "\r\n")
	msg.WriteString("Content-Type: text/html\r\n")

	msg.WriteString("\r\n")
	msg.WriteString(m.Body)

	err := deliverer.SendMail(m.From, m.To, m.Subject, []byte(m.Body))

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
