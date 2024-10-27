// mail.go
package mail

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"time"

	gomail "github.com/wneessen/go-mail"
)

var (
	ErrNoContent = errors.New("email must have either plain text or HTML body")
	ErrNoSubject = errors.New("email must have a subject")
)

// Config holds the mailer configuration
type Config struct {
	Host          string
	Port          int
	Username      string
	Password      string
	From          string
	TemplateFS    fs.FS
	RetryCount    int
	RetryDelay    time.Duration
	HTMLProcessor HTMLProcessor
}

// HTMLProcessor defines the interface for processing HTML content
type HTMLProcessor interface {
	Process(html string) (string, error)
}

// DefaultHTMLProcessor provides a pass-through implementation
type DefaultHTMLProcessor struct{}

func (p *DefaultHTMLProcessor) Process(html string) (string, error) {
	return html, nil
}

// EmailMessage represents the content and recipients of an email
type EmailMessage struct {
	To           []string
	Template     string
	TemplateData interface{}
	Attachments  []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	Data        io.Reader
	ContentType gomail.ContentType
}

// Mailer handles email sending operations
type Mailer struct {
	config        *Config
	client        *gomail.Client
	htmlProcessor HTMLProcessor
}

// New creates a new Mailer instance
func New(cfg *Config) (*Mailer, error) {
	if cfg.RetryCount == 0 {
		cfg.RetryCount = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 2 * time.Second
	}
	if cfg.HTMLProcessor == nil {
		cfg.HTMLProcessor = &DefaultHTMLProcessor{}
	}

	client, err := gomail.NewClient(
		cfg.Host,
		gomail.WithPort(cfg.Port),
		gomail.WithUsername(cfg.Username),
		gomail.WithPassword(cfg.Password),
		gomail.WithTLSPolicy(gomail.TLSOpportunistic),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mail client: %w", err)
	}

	return &Mailer{
		config:        cfg,
		client:        client,
		htmlProcessor: cfg.HTMLProcessor,
	}, nil
}

// Send sends an email using the provided template and data
func (m *Mailer) Send(msg *EmailMessage) error {
	email := gomail.NewMsg()

	if err := m.setAddresses(email, msg.To); err != nil {
		return err
	}

	if err := m.processTemplates(email, msg); err != nil {
		return err
	}

	if err := m.addAttachments(email, msg.Attachments); err != nil {
		return err
	}

	return m.sendWithRetry(email)
}

func (m *Mailer) setAddresses(email *gomail.Msg, to []string) error {
	if err := email.From(m.config.From); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}

	if err := email.To(to...); err != nil {
		return fmt.Errorf("failed to set to addresses: %w", err)
	}

	return nil
}

func (m *Mailer) processTemplates(email *gomail.Msg, msg *EmailMessage) error {
	tmpl, err := template.New("").ParseFS(m.config.TemplateFS, msg.Template)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Process subject
	subject, err := m.executeTemplate(tmpl, "subject", msg.TemplateData)
	if err != nil {
		return fmt.Errorf("failed to execute subject template: %w", err)
	}
	if subject.Len() == 0 {
		return ErrNoSubject
	}
	email.Subject(subject.String())

	// Process bodies
	plainBody, htmlBody, err := m.processBodies(tmpl, msg.TemplateData)
	if err != nil {
		return err
	}

	if plainBody.Len() == 0 && htmlBody.Len() == 0 {
		return ErrNoContent
	}

	return m.setBodies(email, plainBody, htmlBody)
}

func (m *Mailer) executeTemplate(tmpl *template.Template, name string, data interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, err
	}
	return &buf, nil
}

func (m *Mailer) processBodies(tmpl *template.Template, data interface{}) (*bytes.Buffer, *bytes.Buffer, error) {
	plainBody, err := m.executeTemplate(tmpl, "plainBody", data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to execute plain body template: %w", err)
	}

	var htmlBody *bytes.Buffer
	if t := tmpl.Lookup("htmlBody"); t != nil {
		htmlBuf, err := m.executeTemplate(tmpl, "htmlBody", data)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to execute HTML template: %w", err)
		}

		processed, err := m.htmlProcessor.Process(htmlBuf.String())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to process HTML: %w", err)
		}

		htmlBody = &bytes.Buffer{}
		htmlBody.WriteString(processed)
	}

	return plainBody, htmlBody, nil
}

func (m *Mailer) setBodies(email *gomail.Msg, plainBody, htmlBody *bytes.Buffer) error {
	if plainBody.Len() > 0 {
		email.SetBodyString(gomail.TypeTextPlain, plainBody.String())
	}

	if htmlBody.Len() > 0 {
		if plainBody.Len() > 0 {
			email.AddAlternativeString(gomail.TypeTextHTML, htmlBody.String())
		} else {
			email.SetBodyString(gomail.TypeTextHTML, htmlBody.String())
		}
	}

	return nil
}

func (m *Mailer) addAttachments(email *gomail.Msg, attachments []Attachment) error {
	for _, att := range attachments {
		if err := email.AttachReader(
			att.Filename,
			att.Data,
			gomail.WithFileContentType(att.ContentType),
		); err != nil {
			return fmt.Errorf("failed to attach file %s: %w", att.Filename, err)
		}
	}
	return nil
}

func (m *Mailer) sendWithRetry(email *gomail.Msg) error {
	var lastErr error
	for i := 0; i < m.config.RetryCount; i++ {
		if err := m.client.DialAndSend(email); err != nil {
			lastErr = err
			if i < m.config.RetryCount-1 {
				time.Sleep(m.config.RetryDelay)
				continue
			}
		} else {
			return nil
		}
	}
	return fmt.Errorf("failed to send email after %d attempts: %w", m.config.RetryCount, lastErr)
}
