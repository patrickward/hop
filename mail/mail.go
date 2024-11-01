// Package mail provides a simple mailer that sends emails using the SMTP protocol.
package mail

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"strings"
	"time"

	gomail "github.com/wneessen/go-mail"
)

var (
	ErrNoContent = errors.New("email must have either plain text or HTML body")
	ErrNoSubject = errors.New("email must have a subject")
)

// Config holds the mailer configuration
type Config struct {
	// SMTP server configuration
	Host      string // SMTP server host
	Port      int    // SMTP server port
	Username  string // SMTP server username
	Password  string // SMTP server password
	From      string // From address
	AuthType  string // Type of SMTP authentication (see the go-mail package for options). Default is LOGIN.
	TLSPolicy int    // TLS policy for the SMTP connection (see the go-mail package for options). Default is opportunistic.

	// Template configuration
	TemplateFS   fs.FS  // File system for templates
	TemplatePath string // Path to the templates directory in the file system

	// Retry configuration
	RetryCount int           // Number of retry attempts for sending email
	RetryDelay time.Duration // Delay between retry attempts

	// HTML processor for processing HTML content
	HTMLProcessor HTMLProcessor // HTML processor for processing HTML content

	// Company/Branding
	BaseURL        string // Base URL of the website
	CompanyAddress string // Company address
	CompanyName    string // Company name
	LogoURL        string // URL to the company logo
	SupportEmail   string // Support email address
	WebsiteName    string // Name of the website
	WebsiteURL     string // URL to the company website.

	// Links
	SiteLinks        map[string]string // Site links
	SocialMediaLinks map[string]string // Social media links
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

// StringList is an alias for a slice of strings
type StringList = []string

// EmailMessage represents the content and recipients of an email
type EmailMessage struct {
	To           StringList   // List of recipient email addresses
	Templates    StringList   // List of template names to proccess
	TemplateData any          // Data to be passed to the templates
	Attachments  []Attachment // List of attachments
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

// NewMailer creates a new Mailer instance
func NewMailer(cfg *Config) (*Mailer, error) {
	if cfg.RetryCount == 0 {
		cfg.RetryCount = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 2 * time.Second
	}
	if cfg.HTMLProcessor == nil {
		cfg.HTMLProcessor = &DefaultHTMLProcessor{}
	}

	authType := authTypeFromString(cfg.AuthType)
	tlsPolicy := tlsPolicyFromInt(cfg.TLSPolicy)

	client, err := gomail.NewClient(
		cfg.Host,
		gomail.WithTimeout(10*time.Second),
		gomail.WithSMTPAuth(authType),
		gomail.WithPort(cfg.Port),
		gomail.WithUsername(cfg.Username),
		gomail.WithPassword(cfg.Password),
		gomail.WithTLSPolicy(tlsPolicy),
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

// Config returns the mailer configuration
func (m *Mailer) Config() *Config {
	return m.config
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

// NewTemplateData creates a new template data map with default values
func (m *Mailer) NewTemplateData() TemplateData {
	return NewTemplateData(m.config)
}

func (m *Mailer) processTemplates(email *gomail.Msg, msg *EmailMessage) error {
	templatePath := msg.Templates
	if m.config.TemplatePath != "" {
		// For each template, we need to prepend the template path
		for i, tmpl := range msg.Templates {
			msg.Templates[i] = strings.TrimSuffix(m.config.TemplatePath, "/") + "/" + tmpl
		}
	}

	tmpl, err := template.New("").ParseFS(m.config.TemplateFS, templatePath...)
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

func (m *Mailer) executeTemplate(tmpl *template.Template, name string, data any) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, err
	}
	return &buf, nil
}

func (m *Mailer) processBodies(tmpl *template.Template, data any) (*bytes.Buffer, *bytes.Buffer, error) {
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

func authTypeFromString(typ string) gomail.SMTPAuthType {
	switch typ {
	case "PLAIN":
		return gomail.SMTPAuthPlain
	case "LOGIN":
		return gomail.SMTPAuthLogin
	case "CRAM-MD5":
		return gomail.SMTPAuthCramMD5
	case "NOAUTH":
		return gomail.SMTPAuthNoAuth
	case "XOAUTH2":
		return gomail.SMTPAuthXOAUTH2
	case "CUSTOM":
		return gomail.SMTPAuthCustom
	case "SCRAM-SHA-1":
		return gomail.SMTPAuthSCRAMSHA1
	case "SCRAM-SHA-1-PLUS":
		return gomail.SMTPAuthSCRAMSHA1PLUS
	case "SCRAM-SHA-256":
		return gomail.SMTPAuthSCRAMSHA256
	case "SCRAM-SHA-256-PLUS":
		return gomail.SMTPAuthSCRAMSHA256PLUS
	default:
		return gomail.SMTPAuthLogin
	}
}

func tlsPolicyFromInt(typ int) gomail.TLSPolicy {
	switch typ {
	case 0:
		return gomail.NoTLS
	case 1:
		return gomail.TLSOpportunistic
	case 2:
		return gomail.TLSMandatory
	default:
		return gomail.TLSOpportunistic
	}
}
