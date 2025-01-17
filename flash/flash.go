package flash

import (
	"context"
	"encoding/gob"

	"github.com/alexedwards/scs/v2"
)

const SessionKey = "hop_session_flash"

// Messages is a slice of flash messages.
type Messages []Message

// MessageType represents the type of flash message.
type MessageType string

const (
	// MessageTypeError is an error message
	MessageTypeError MessageType = "error"
	// MessageTypeSuccess is a success message
	MessageTypeSuccess MessageType = "success"
	// MessageTypeWarning is a warning message
	MessageTypeWarning MessageType = "warning"
	// MessageTypeInfo is an informational message
	MessageTypeInfo MessageType = "info"
)

// Message represents a flash message.
type Message struct {
	Type      MessageType // The type of message
	Content   string      // The message to display
	Closable  bool        // Closable by default; this flag makes it not closable
	AutoClose int         // Duration in milliseconds
}

// NotClosable returns true if the flash message is not closable.
func (f Message) NotClosable() bool {
	return !f.Closable
}

func init() {
	gob.Register(Message{})
	gob.Register(Messages{})
}

// Manager manages flash messages.
type Manager struct {
	session *scs.SessionManager
}

// NewManager creates a new Manager instance.
func NewManager(session *scs.SessionManager) *Manager {
	return &Manager{session: session}
}

// Info adds an informational flash message.
func (m *Manager) Info(ctx context.Context, content string, closeable bool, autoclose int) {
	m.add(ctx, Message{Type: MessageTypeInfo, Content: content, Closable: closeable, AutoClose: autoclose})
}

// Error adds an error flash message.
func (m *Manager) Error(ctx context.Context, content string, closeable bool, autoclose int) {
	m.add(ctx, Message{Type: MessageTypeError, Content: content, Closable: closeable, AutoClose: autoclose})
}

// Success adds a success flash message.
func (m *Manager) Success(ctx context.Context, content string, closeable bool, autoclose int) {
	m.add(ctx, Message{Type: MessageTypeSuccess, Content: content, Closable: closeable, AutoClose: autoclose})
}

// Warning adds a warning flash message.
func (m *Manager) Warning(ctx context.Context, content string, closeable bool, autoclose int) {
	m.add(ctx, Message{Type: MessageTypeWarning, Content: content, Closable: closeable, AutoClose: autoclose})
}

// NewInfoMessage creates a new informational flash message, but does not add it to the session.
func (m *Manager) NewInfoMessage(content string, closeable bool, autoclose int) Message {
	return Message{Type: MessageTypeInfo, Content: content, Closable: closeable, AutoClose: autoclose}
}

// NewErrorMessage creates a new error flash message, but does not add it to the session.
func (m *Manager) NewErrorMessage(content string, closeable bool, autoclose int) Message {
	return Message{Type: MessageTypeError, Content: content, Closable: closeable, AutoClose: autoclose}
}

// NewSuccessMessage creates a new success flash message, but does not add it to the session.
func (m *Manager) NewSuccessMessage(content string, closeable bool, autoclose int) Message {
	return Message{Type: MessageTypeSuccess, Content: content, Closable: closeable, AutoClose: autoclose}
}

// NewWarningMessage creates a new warning flash message, but does not add it to the session.
func (m *Manager) NewWarningMessage(content string, closeable bool, autoclose int) Message {
	return Message{Type: MessageTypeWarning, Content: content, Closable: closeable, AutoClose: autoclose}
}

// AsMessages ensures we have a slice of messages by converting a single message to a slice if necessary
func (m *Manager) AsMessages(msg Message) []Message {
	return []Message{msg}
}

// Get returns the flash messages from the session.
func (m *Manager) Get(ctx context.Context) Messages {
	var flashes Messages

	if v := m.session.Pop(ctx, SessionKey); v != nil {
		flashes = v.(Messages)
	}

	return flashes
}

// ReplaceFlash replaces the flash messages in the session.
func (m *Manager) ReplaceFlash(ctx context.Context, flashes Messages) {
	m.session.Put(ctx, SessionKey, flashes)
}

func (m *Manager) add(ctx context.Context, msg Message) {
	var flashes Messages

	if v := m.session.Get(ctx, SessionKey); v == nil {
		flashes = Messages{}
	} else {
		flashes = v.(Messages)
	}

	flashes = append(flashes, msg)
	m.session.Put(ctx, SessionKey, flashes)
}
