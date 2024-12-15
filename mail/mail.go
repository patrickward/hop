// Package mail provides a simple mailer that sends emails using the SMTP protocol.
package mail

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"strings"
	"time"

	gomail "github.com/wneessen/go-mail"

	"github.com/patrickward/hop/templates"
)

var (
	ErrNoContent = errors.New("email must have either plain text or HTML body")
	ErrNoSubject = errors.New("email must have a subject")
)

// SMTPClient defines the interface for an SMTP client, mainly used for testing
type SMTPClient interface {
	DialAndSend(messages ...*gomail.Msg) error
}

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
	TemplateFS      fs.FS            // File system for templates
	TemplatePath    string           // Path to the templates directory in the file system
	TemplateFuncMap template.FuncMap // Template function map that gets merged with the default function map from render

	// Retry configuration
	RetryCount int           // Number of retry attempts for sending email
	RetryDelay time.Duration // Delay between retry attempts

	// HTML processor for processing HTML content
	HTMLProcessor HTMLProcessor // HTML processor for processing HTML content

	// Company/Branding
	BaseURL         string // Base URL of the website
	CompanyAddress1 string // The first line of the company address (usually the street address)
	CompanyAddress2 string // The second line of the company address (usually the city, state, and ZIP code)
	CompanyName     string // Company name
	LogoURL         string // URL to the company logo
	SupportEmail    string // Support email address
	SupportPhone    string // Support phone number
	WebsiteName     string // Name of the website
	WebsiteURL      string // URL to the company website.

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

// Mailer handles email sending operations
type Mailer struct {
	config *Config
	//client        *gomail.Client
	client        SMTPClient
	funcMap       template.FuncMap
	htmlProcessor HTMLProcessor
}

// NewMailer creates a new Mailer instance using the provided configuration and the default SMTP client
func NewMailer(cfg *Config) (*Mailer, error) {
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

	return NewMailerWithClient(cfg, client), nil
}

// NewMailerWithClient creates a new Mailer with a provided SMTP client
func NewMailerWithClient(cfg *Config, client SMTPClient) *Mailer {
	if cfg.RetryCount == 0 {
		cfg.RetryCount = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 2 * time.Second
	}
	if cfg.HTMLProcessor == nil {
		cfg.HTMLProcessor = &DefaultHTMLProcessor{}
	}

	//funcMap := render.MergeFuncMaps(cfg.TemplateFuncMap)
	funcMap := templates.MergeFuncMaps(templates.FuncMap(), cfg.TemplateFuncMap)

	return &Mailer{
		config:        cfg,
		client:        client,
		funcMap:       funcMap,
		htmlProcessor: cfg.HTMLProcessor,
	}
}

// Config returns the mailer configuration
func (m *Mailer) Config() *Config {
	return m.config
}

// Send sends an email using the provided template and data
func (m *Mailer) Send(msg *Message) error {
	email := gomail.NewMsg()

	if err := m.setAddresses(email, msg); err != nil {
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

// setAddresses sets all address fields on the email
func (m *Mailer) setAddresses(email *gomail.Msg, msg *Message) error {
	// Set From address
	if err := email.From(m.config.From); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}

	// Set To addresses
	if err := email.To(msg.To...); err != nil {
		return fmt.Errorf("failed to set to addresses: %w", err)
	}

	// Set CC addresses if present
	if len(msg.Cc) > 0 {
		if err := email.Cc(msg.Cc...); err != nil {
			return fmt.Errorf("failed to set cc addresses: %w", err)
		}
	}

	// Set BCC addresses if present
	if len(msg.Bcc) > 0 {
		if err := email.Bcc(msg.Bcc...); err != nil {
			return fmt.Errorf("failed to set bcc addresses: %w", err)
		}
	}

	// Set Reply-To if present
	if msg.ReplyTo != "" {
		if err := email.ReplyTo(msg.ReplyTo); err != nil {
			return fmt.Errorf("failed to set reply-to address: %w", err)
		}
	}

	return nil
}

// NewTemplateData creates a new template data map with default values
func (m *Mailer) NewTemplateData() TemplateData {
	return NewTemplateData(m.config)
}

func (m *Mailer) processTemplates(email *gomail.Msg, msg *Message) error {
	templatePath := msg.Templates
	if m.config.TemplatePath != "" {
		// For each template, we need to prepend the template path
		for i, tmpl := range msg.Templates {
			msg.Templates[i] = strings.TrimSuffix(m.config.TemplatePath, "/") + "/" + tmpl
		}
	}

	tmpl, err := template.New("").Funcs(m.funcMap).ParseFS(m.config.TemplateFS, templatePath...)
	if err != nil {
		if templatePath == nil {
			templatePath = []string{""}
		}
		return &TemplateError{
			TemplateName: templatePath[0],
			OriginalErr:  err,
			Phase:        "parse",
		}
	}

	// Process subject
	subject, err := m.executeTemplate(tmpl, "subject", msg.TemplateData)
	if err != nil {
		return &TemplateError{
			TemplateName: "subject",
			OriginalErr:  err,
			Phase:        "execute",
		}
	}
	if subject.Len() == 0 {
		return ErrNoSubject
	}
	email.Subject(subject.String())

	// Process bodies
	textPlain, textHTML, err := m.processBodies(tmpl, msg.TemplateData)
	if err != nil {
		return err
	}

	if textPlain.Len() == 0 && textHTML.Len() == 0 {
		return ErrNoContent
	}

	return m.setBodies(email, textPlain, textHTML)
}

func (m *Mailer) executeTemplate(tmpl *template.Template, name string, data any) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, err
	}
	return &buf, nil
}

// Template helper methods for Mailer
func (m *Mailer) processBodies(tmpl *template.Template, data any) (*bytes.Buffer, *bytes.Buffer, error) {
	// Execute plain body template
	textPlain, err := m.executeTemplate(tmpl, "text/plain", data)
	if err != nil {
		return nil, nil, &TemplateError{
			TemplateName: "text/plain",
			OriginalErr:  err,
			Phase:        "execute",
		}
	}

	// Execute HTML body template if it exists
	var textHTML *bytes.Buffer
	if t := tmpl.Lookup("text/html"); t != nil {
		htmlBuf, err := m.executeTemplate(tmpl, "text/html", data)
		if err != nil {
			return nil, nil, &TemplateError{
				TemplateName: "text/html",
				OriginalErr:  err,
				Phase:        "execute",
			}
		}

		// Process HTML if we have a processor
		if m.htmlProcessor != nil {
			processed, err := m.htmlProcessor.Process(htmlBuf.String())
			if err != nil {
				return nil, nil, &TemplateError{
					TemplateName: "text/html",
					OriginalErr:  err,
					Phase:        "process",
				}
			}
			textHTML = bytes.NewBufferString(processed)
		} else {
			textHTML = htmlBuf
		}
	}

	return textPlain, textHTML, nil
}

func (m *Mailer) setBodies(email *gomail.Msg, textPlain, textHTML *bytes.Buffer) error {
	if textPlain.Len() > 0 {
		email.SetBodyString(gomail.TypeTextPlain, textPlain.String())
	}

	if textHTML.Len() > 0 {
		if textPlain.Len() > 0 {
			email.AddAlternativeString(gomail.TypeTextHTML, textHTML.String())
		} else {
			email.SetBodyString(gomail.TypeTextHTML, textHTML.String())
		}
	}

	return nil
}

func (m *Mailer) addAttachments(email *gomail.Msg, attachments []Attachment) error {
	for _, att := range attachments {
		var opts []gomail.FileOption
		if att.ContentType != "" {
			opts = append(opts, gomail.WithFileContentType(att.ContentType))
		}

		if att.Data == nil {
			return fmt.Errorf("nil reader for attachment %s", att.Filename)
		}

		if err := email.AttachReader(att.Filename, att.Data, opts...); err != nil {
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
