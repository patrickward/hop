package mail_test

import (
	"embed"
	"os"
	"strings"
	"testing"
	"time"

	gomail "github.com/wneessen/go-mail"

	"github.com/patrickward/hop/internal/testutil"
	"github.com/patrickward/hop/mail"
)

//go:embed testdata/*
var testFS embed.FS

type mockHTMLProcessor struct {
	processFunc func(string) (string, error)
}

func (m *mockHTMLProcessor) Process(html string) (string, error) {
	if m.processFunc != nil {
		return m.processFunc(html)
	}
	return html, nil
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Exit with test result code
	os.Exit(code)
}

func TestMailer(t *testing.T) {
	cleanup := testutil.SetupMailpit(t)
	defer cleanup()

	cfg := &mail.Config{
		Host:          "localhost",
		Port:          1025,
		From:          "test@example.com",
		AuthType:      string(gomail.SMTPAuthNoAuth),
		TemplateFS:    testFS,
		RetryCount:    1,
		RetryDelay:    time.Millisecond,
		HTMLProcessor: &mockHTMLProcessor{},
	}

	mailer, err := mail.NewMailer(cfg)
	if err != nil {
		t.Fatalf("Failed to create mailer: %v", err)
	}

	tests := []struct {
		name    string
		msg     *mail.EmailMessage
		wantErr bool
		check   func(*testing.T, testutil.MailpitMessage)
	}{
		{
			name: "basic email with both bodies",
			msg: &mail.EmailMessage{
				To:        mail.StringList{"recipient@example.com"},
				Templates: mail.StringList{"testdata/basic.tmpl"},
				TemplateData: map[string]string{
					"Name": "John",
				},
			},
			wantErr: false,
			check: func(t *testing.T, msg testutil.MailpitMessage) {
				if msg.Subject != "Test Email" {
					t.Errorf("Wrong subject: got %v, want %v",
						msg.Subject, "Test Email")
				}
				if len(msg.To) != 1 || msg.To[0].Address != "recipient@example.com" {
					t.Errorf("Wrong recipient: got %v, want %v",
						msg.To[0].Address, "recipient@example.com")
				}
				if msg.From.Address != "test@example.com" {
					t.Errorf("Wrong sender: got %v, want %v",
						msg.From.Address, "test@example.com")
				}
			},
		},

		// Add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.ClearMailpitMessages(t)

			err := mailer.Send(tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v; make sure mailpit is running!", err, tt.wantErr)
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

// Additional test cases for error scenarios
func TestMailer_Errors(t *testing.T) {
	cleanup := testutil.SetupMailpit(t)
	defer cleanup()

	tests := []struct {
		name    string
		cfg     *mail.Config
		msg     *mail.EmailMessage
		wantErr string
	}{
		{
			name: "invalid port",
			cfg: &mail.Config{
				Host:       "localhost",
				Port:       1234, // Wrong port
				From:       "test@example.com",
				TemplateFS: testFS,
			},
			msg: &mail.EmailMessage{
				To:        mail.StringList{"test@example.com"},
				Templates: mail.StringList{"testdata/welcome.tmpl"},
				TemplateData: map[string]interface{}{
					"Name":    "Test User",
					"Company": "Test Co",
				},
			},
			wantErr: "connection refused",
		},
		{
			name: "invalid template path",
			cfg: &mail.Config{
				Host:       "localhost",
				Port:       1025,
				From:       "test@example.com",
				TemplateFS: testFS,
			},
			msg: &mail.EmailMessage{
				To:        mail.StringList{"test@example.com"},
				Templates: mail.StringList{"testdata/nonexistent.tmpl"},
				TemplateData: map[string]interface{}{
					"Name":    "Test User",
					"Company": "Test Co",
				},
			},
			wantErr: "failed to parse template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer, err := mail.NewMailer(tt.cfg)
			if err != nil {
				t.Fatalf("Failed to create mailer: %v", err)
			}

			err = mailer.Send(tt.msg)
			if err == nil {
				t.Error("Expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}
