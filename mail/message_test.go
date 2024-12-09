package mail_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/patrickward/hop/mail"
)

func TestMessageBuilder(t *testing.T) {
	tests := []struct {
		name      string
		build     func(*mail.Builder)
		wantErr   bool
		errString string
		validate  func(*testing.T, *mail.Message)
	}{
		{
			name: "basic message",
			build: func(b *mail.Builder) {
				b.To("user@example.com").
					Template("welcome.tmpl").
					WithData(map[string]string{"name": "John"})
			},
			validate: func(t *testing.T, msg *mail.Message) {
				assert.Equal(t, []string{"user@example.com"}, msg.To)
				assert.Equal(t, []string{"welcome.tmpl"}, msg.Templates)
				assert.Equal(t, map[string]string{"name": "John"}, msg.TemplateData)
			},
		},
		{
			name: "message with cc and bcc",
			build: func(b *mail.Builder) {
				b.To("user@example.com").
					Cc("cc@example.com").
					Bcc("bcc@example.com").
					Template("notify.tmpl")
			},
			validate: func(t *testing.T, msg *mail.Message) {
				assert.Equal(t, []string{"user@example.com"}, msg.To)
				assert.Equal(t, []string{"cc@example.com"}, msg.Cc)
				assert.Equal(t, []string{"bcc@example.com"}, msg.Bcc)
			},
		},
		{
			name: "message with reply-to",
			build: func(b *mail.Builder) {
				b.To("user@example.com").
					ReplyTo("reply@example.com").
					Template("notify.tmpl")
			},
			validate: func(t *testing.T, msg *mail.Message) {
				assert.Equal(t, []string{"user@example.com"}, msg.To)
				assert.Equal(t, "reply@example.com", msg.ReplyTo)
			},
		},
		{
			name: "missing recipient",
			build: func(b *mail.Builder) {
				b.Template("welcome.tmpl")
			},
			wantErr:   true,
			errString: "email must have at least one recipient",
		},
		{
			name: "missing template",
			build: func(b *mail.Builder) {
				b.To("user@example.com")
			},
			wantErr:   true,
			errString: "email must have at least one template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := mail.NewMessage()
			tt.build(b)
			msg, err := b.Build()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, msg)
			if tt.validate != nil {
				tt.validate(t, msg)
			}
		})
	}
}
