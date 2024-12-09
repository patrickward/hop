package mail_test

import (
	"bytes"
	"embed"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomail "github.com/wneessen/go-mail"

	"github.com/patrickward/hop/mail"
)

//go:embed testdata/*
var testFS embed.FS

func testConfig() *mail.Config {
	return &mail.Config{
		Host:       "localhost",
		Port:       1025,
		From:       "test@example.com",
		AuthType:   string(gomail.SMTPAuthNoAuth),
		TemplateFS: testFS,
		RetryCount: 1,
		RetryDelay: time.Millisecond,
	}
}

func TestMailer_Send(t *testing.T) {
	tests := []struct {
		name      string
		config    *mail.Config
		buildMsg  func() (*mail.Message, error)
		setupMock func(*mockSMTPClient)
		wantErr   bool
		errString string
		validate  func(*testing.T, mockMessage)
	}{
		{
			name:   "basic email",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/basic.tmpl").
					WithData(map[string]string{"name": "John"}).
					Build()
			},
			validate: func(t *testing.T, msg mockMessage) {
				// Validate From address
				require.Len(t, msg.from, 1)
				assert.Equal(t, "test@example.com", msg.from[0].Address)

				// Validate To address
				require.Len(t, msg.to, 1)
				assert.Equal(t, "recipient@example.com", msg.to[0].Address)

				// Validate subject from template
				assert.Equal(t, "Test Email", msg.subject)
			},
		},
		{
			name:   "with cc and bcc",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient@example.com").
					Cc("cc@example.com").
					Bcc("bcc@example.com").
					Template("testdata/basic.tmpl").
					WithData(map[string]string{"name": "John"}).
					Build()
			},
			validate: func(t *testing.T, msg mockMessage) {
				require.Len(t, msg.cc, 1)
				assert.Equal(t, "cc@example.com", msg.cc[0].Address)
				require.Len(t, msg.bcc, 1)
				assert.Equal(t, "bcc@example.com", msg.bcc[0].Address)
			},
		},
		{
			name:   "with reply-to",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient@example.com").
					ReplyTo("reply@example.com").
					Template("testdata/basic.tmpl").
					WithData(map[string]string{"name": "John"}).
					Build()
			},
			validate: func(t *testing.T, msg mockMessage) {
				assert.Equal(t, "<reply@example.com>", msg.replyTo)
			},
		},
		{
			name:   "smtp error",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/basic.tmpl").
					Build()
			},
			setupMock: func(m *mockSMTPClient) {
				m.SetError("smtp connection failed")
			},
			wantErr:   true,
			errString: "smtp connection failed",
		},
		{
			name: "custom from address",
			config: func() *mail.Config {
				cfg := testConfig()
				cfg.From = "custom@example.com"
				return cfg
			}(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/basic.tmpl").
					Build()
			},
			validate: func(t *testing.T, msg mockMessage) {
				require.Len(t, msg.from, 1)
				assert.Equal(t, "custom@example.com", msg.from[0].Address)
			},
		},
		{
			name:   "with multiple recipients",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient1@example.com", "recipient2@example.com").
					Template("testdata/basic.tmpl").
					Build()
			},
			validate: func(t *testing.T, msg mockMessage) {
				require.Len(t, msg.to, 2)
				assert.Equal(t, "recipient1@example.com", msg.to[0].Address)
				assert.Equal(t, "recipient2@example.com", msg.to[1].Address)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockSMTPClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			mailer := mail.NewMailerWithClient(tt.config, mock)

			msg, err := tt.buildMsg()
			require.NoError(t, err)

			err = mailer.Send(msg)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
				return
			}

			require.NoError(t, err)
			require.Len(t, mock.sentMessages, 1)

			if tt.validate != nil {
				tt.validate(t, mock.sentMessages[0])
			}
		})
	}
}

func TestTemplateProcessing(t *testing.T) {
	tests := []struct {
		name      string
		config    *mail.Config
		buildMsg  func() (*mail.Message, error)
		setupMock func(*mockSMTPClient)
		wantErr   bool
		errString string
		validate  func(*testing.T, mockMessage)
	}{
		{
			name:   "template with subject and body",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/basic.tmpl").
					WithData(map[string]string{
						"name": "John",
					}).
					Build()
			},
			validate: func(t *testing.T, msg mockMessage) {
				assert.Equal(t, "Test Email", msg.subject)
			},
		},
		{
			name:   "missing template",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/nonexistent.tmpl").
					Build()
			},
			wantErr:   true,
			errString: "template error",
		},
		{
			name:   "template with missing section",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/missing_subject.tmpl").
					Build()
			},
			wantErr:   true,
			errString: "template error",
		},
		{
			name:   "multiple templates",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient@example.com").
					Template(
						"testdata/with_header.tmpl",
						"testdata/header.tmpl",
					).
					WithData(map[string]string{
						"name": "John",
					}).
					Build()
			},
			validate: func(t *testing.T, msg mockMessage) {
				assert.Equal(t, "Welcome, John", msg.subject)
				assert.Contains(t, msg.bodyHTML, "Header for John")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockSMTPClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			mailer := mail.NewMailerWithClient(tt.config, mock)

			msg, err := tt.buildMsg()
			require.NoError(t, err)

			err = mailer.Send(msg)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
				return
			}

			require.NoError(t, err)
			require.Len(t, mock.sentMessages, 1)

			if tt.validate != nil {
				tt.validate(t, mock.sentMessages[0])
			}
		})
	}
}

func TestAttachments(t *testing.T) {
	tests := []struct {
		name      string
		config    *mail.Config
		buildMsg  func() (*mail.Message, error)
		setupMock func(*mockSMTPClient)
		wantErr   bool
		errString string
		validate  func(*testing.T, mockMessage)
	}{
		{
			name:   "single attachment from reader",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				content := bytes.NewReader([]byte("test content"))
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/basic.tmpl").
					Attach("test.txt", content).
					Build()
			},
		},
		{
			name:   "multiple attachments",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				content1 := bytes.NewReader([]byte("content 1"))
				content2 := bytes.NewReader([]byte("content 2"))
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/basic.tmpl").
					Attach("file1.txt", content1).
					Attach("file2.txt", content2).
					Build()
			},
		},
		{
			name:   "attachment with custom content type",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				content := bytes.NewReader([]byte("<html><body>Test</body></html>"))
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/basic.tmpl").
					AttachWithContentType("test.html", content, gomail.TypeTextHTML).
					Build()
			},
		},
		{
			name:   "attachment with empty reader",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/basic.tmpl").
					Attach("empty.txt", strings.NewReader("")).
					Build()
			},
		},
		{
			name:   "attachment with nil reader",
			config: testConfig(),
			buildMsg: func() (*mail.Message, error) {
				var nilReader io.Reader
				return mail.NewMessage().
					To("recipient@example.com").
					Template("testdata/basic.tmpl").
					Attach("nil.txt", nilReader).
					Build()
			},
			wantErr:   true,
			errString: "nil reader",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockSMTPClient()
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			mailer := mail.NewMailerWithClient(tt.config, mock)

			msg, err := tt.buildMsg()
			require.NoError(t, err)

			err = mailer.Send(msg)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
				return
			}

			require.NoError(t, err)
			require.Len(t, mock.sentMessages, 1)

			if tt.validate != nil {
				tt.validate(t, mock.sentMessages[0])
			}
		})
	}
}

func TestOpenFileAttachment(t *testing.T) {
	tests := []struct {
		name         string
		setupFile    func() string
		wantErr      bool
		errString    string
		validateFile func(*testing.T, string, io.Reader, func() error)
	}{
		{
			name: "valid file",
			setupFile: func() string {
				f, err := os.CreateTemp("", "mailTest-*.txt")
				require.NoError(t, err)
				_, err = f.WriteString("test content")
				require.NoError(t, err)
				require.NoError(t, f.Close())
				return f.Name()
			},
			validateFile: func(t *testing.T, filename string, reader io.Reader, cleanup func() error) {
				defer func() {
					require.NoError(t, cleanup())
				}()

				assert.Equal(t, "mailTest", strings.Split(path.Base(filename), "-")[0])

				content, err := io.ReadAll(reader)
				require.NoError(t, err)
				assert.Equal(t, "test content", string(content))
			},
		},
		{
			name: "nonexistent file",
			setupFile: func() string {
				return "nonexistent.txt"
			},
			wantErr:   true,
			errString: "no such file",
		},
		{
			name: "directory instead of file",
			setupFile: func() string {
				dir, err := os.MkdirTemp("", "mail-test-*")
				require.NoError(t, err)
				return dir
			},
			wantErr:   true,
			errString: "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filepath := tt.setupFile()
			if !strings.HasPrefix(filepath, "/tmp") && !strings.HasPrefix(filepath, "nonexistent") {
				defer func(path string) {
					_ = os.RemoveAll(path)
				}(filepath)
			}

			filename, reader, cleanup, err := mail.OpenFileAttachment(filepath)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, reader)
			require.NotNil(t, cleanup)

			if tt.validateFile != nil {
				tt.validateFile(t, filename, reader, cleanup)
			}
		})
	}
}
