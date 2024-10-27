package notifications_test

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/patrickward/hop/internal/testutil"
	"github.com/patrickward/hop/mail"
	"github.com/patrickward/hop/mail/notifications"
)

//go:embed templates/*
var testFS embed.FS

// TestWelcomeEmailSender_Send tests the welcome email functionality
func TestWelcomeEmailSender_Send(t *testing.T) {
	testutil.CheckDockerAvailable(t)
	cleanup := testutil.SetupMailpit(t)
	defer cleanup()

	// Create a temporary directory for test guides
	tempDir, err := os.MkdirTemp("", "welcome-guides-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test PDF guide
	guideContent := []byte("%PDF-1.4\nTest guide content")
	if err := os.WriteFile(
		filepath.Join(tempDir, "enterprise.pdf"),
		guideContent,
		0644,
	); err != nil {
		t.Fatalf("Failed to create test guide: %v", err)
	}

	// Create mail client
	cfg := &mail.Config{
		Host:       "localhost",
		Port:       1025,
		From:       "welcome@example.com",
		TemplateFS: testFS,
	}

	mailer, err := mail.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create mailer: %v", err)
	}

	// Create welcome email sender
	sender := notifications.NewWelcomeEmailSender(mailer, tempDir)

	tests := []struct {
		name    string
		to      string
		data    notifications.WelcomeEmailData
		wantErr bool
		check   func(*testing.T, testutil.MailpitMessage)
	}{
		{
			name: "basic welcome email",
			to:   "newuser@example.com",
			data: notifications.WelcomeEmailData{
				Name:    "Alice Smith",
				Company: "TechCorp",
			},
			wantErr: false,
			check: func(t *testing.T, msg testutil.MailpitMessage) {
				if msg.Subject != "Welcome to TechCorp, Alice Smith!" {
					t.Errorf("Wrong subject: got %v, want %v",
						msg.Subject, "Welcome to TechCorp, Alice Smith!")
				}

				if msg.From.Address != "welcome@example.com" {
					t.Errorf("Wrong sender: got %v, want %v",
						msg.From.Address, "welcome@example.com")
				}

				if len(msg.To) != 1 || msg.To[0].Address != "newuser@example.com" {
					t.Errorf("Wrong recipient: got %v, want %v",
						msg.To[0].Address, "newuser@example.com")
				}

				if msg.Attachments != 0 {
					t.Errorf("Expected no attachments, got %d", msg.Attachments)
				}
			},
		},
		{
			name: "welcome email with guide",
			to:   "newuser@example.com",
			data: notifications.WelcomeEmailData{
				Name:    "Bob Johnson",
				Company: "TechCorp",
				GuideID: "enterprise",
			},
			wantErr: false,
			check: func(t *testing.T, msg testutil.MailpitMessage) {
				if msg.Subject != "Welcome to TechCorp, Bob Johnson!" {
					t.Errorf("Wrong subject: got %v, want %v",
						msg.Subject, "Welcome to TechCorp, Bob Johnson!")
				}

				if msg.Attachments != 1 {
					t.Errorf("Expected 1 attachment, got %d", msg.Attachments)
				}
			},
		},
		{
			name: "welcome email with invalid guide",
			to:   "newuser@example.com",
			data: notifications.WelcomeEmailData{
				Name:    "Charlie Brown",
				Company: "TechCorp",
				GuideID: "nonexistent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ClearMailpitMessages(t)

			err := sender.Send(tt.to, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				messages := testutil.GetMailpitMessages(t)
				if len(messages) != 1 {
					t.Errorf("Expected 1 message, got %d", len(messages))
					return
				}

				if tt.check != nil {
					tt.check(t, messages[0])
				}
			}
		})
	}
}
