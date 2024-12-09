package mail_test

import (
	"fmt"
	"net/mail"

	gomail "github.com/wneessen/go-mail"
)

type mockSMTPClient struct {
	sentMessages []mockMessage
	shouldError  bool
	errorMsg     string
}

type mockMessage struct {
	from        []*mail.Address
	to          []*mail.Address
	cc          []*mail.Address
	bcc         []*mail.Address
	replyTo     string
	subject     string
	bodyPlain   string
	bodyHTML    string
	attachments []mockAttachment
}

type mockAttachment struct {
	filename    string
	contentType string
	data        []byte
}

func newMockSMTPClient() *mockSMTPClient {
	return &mockSMTPClient{
		sentMessages: make([]mockMessage, 0),
	}
}

func (m *mockSMTPClient) DialAndSend(messages ...*gomail.Msg) error {
	if m.shouldError {
		if m.errorMsg != "" {
			return fmt.Errorf(m.errorMsg)
		}
		return fmt.Errorf("mock smtp error")
	}

	for _, msg := range messages {
		replyTo := ""
		subject := ""
		if replyTos := msg.GetGenHeader(gomail.HeaderReplyTo); len(replyTos) > 0 {
			replyTo = replyTos[0]
		}
		if subjects := msg.GetGenHeader(gomail.HeaderSubject); len(subjects) > 0 {
			subject = subjects[0]
		}

		mockMsg := mockMessage{
			from:    msg.GetFrom(),
			to:      msg.GetTo(),
			cc:      msg.GetCc(),
			bcc:     msg.GetBcc(),
			replyTo: replyTo,
			subject: subject,
		}

		parts := msg.GetParts()
		for _, part := range parts {
			if part.GetContentType() == "text/plain" {
				data, err := part.GetContent()
				if err != nil {
					return err
				}
				mockMsg.bodyPlain = string(data)
			}
			if part.GetContentType() == "text/html" {
				data, err := part.GetContent()
				if err != nil {
					return err
				}
				mockMsg.bodyHTML = string(data)
			}
		}

		m.sentMessages = append(m.sentMessages, mockMsg)
	}

	return nil
}

// Helper methods for tests
func (m *mockSMTPClient) LastMessage() (mockMessage, error) {
	if len(m.sentMessages) == 0 {
		return mockMessage{}, fmt.Errorf("no messages sent")
	}
	return m.sentMessages[len(m.sentMessages)-1], nil
}

func (m *mockSMTPClient) SetError(err string) {
	m.shouldError = true
	m.errorMsg = err
}

func (m *mockSMTPClient) Reset() {
	m.sentMessages = make([]mockMessage, 0)
	m.shouldError = false
	m.errorMsg = ""
}
