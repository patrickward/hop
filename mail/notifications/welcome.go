package notifications

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/patrickward/hop/mail" // base mail package we created
)

// WelcomeEmailData contains all data needed for welcome email
type WelcomeEmailData struct {
	Name    string
	Company string
	GuideID string // Optional: ID of welcome guide to attach
}

// WelcomeEmailSender handles welcome email specific operations
type WelcomeEmailSender struct {
	mailer     *mail.Mailer
	guidesPath string // path to welcome guides
}

// NewWelcomeEmailSender creates a new welcome email sender
func NewWelcomeEmailSender(mailer *mail.Mailer, guidesPath string) *WelcomeEmailSender {
	return &WelcomeEmailSender{
		mailer:     mailer,
		guidesPath: guidesPath,
	}
}

// Send sends a welcome email to a new user
func (s *WelcomeEmailSender) Send(to string, data WelcomeEmailData) error {
	msg := &mail.EmailMessage{
		To:       []string{to},
		Template: "templates/welcome.tmpl",
		TemplateData: map[string]interface{}{
			"Name":    data.Name,
			"Company": data.Company,
		},
	}

	// Optionally attach welcome guide if GuideID is provided
	if data.GuideID != "" {
		guide, err := s.attachWelcomeGuide(data.GuideID)
		if err != nil {
			return fmt.Errorf("failed to attach welcome guide: %w", err)
		}
		msg.Attachments = []mail.Attachment{guide}
	}

	return s.mailer.Send(msg)
}

func (s *WelcomeEmailSender) attachWelcomeGuide(guideID string) (mail.Attachment, error) {
	guidePath := filepath.Join(s.guidesPath, fmt.Sprintf("%s.pdf", guideID))

	file, err := os.Open(guidePath)
	if err != nil {
		return mail.Attachment{}, fmt.Errorf("failed to open guide file: %w", err)
	}

	return mail.Attachment{
		Filename:    fmt.Sprintf("welcome-guide-%s.pdf", guideID),
		Data:        file,
		ContentType: "application/pdf",
	}, nil
}
