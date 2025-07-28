package alert

import (
	"context"
	"encoding/gob"
)

func init() {
	gob.Register(Message{})
	gob.Register(Messages{})
}

type FlashManager struct {
	key     string         // The key used to store flash messages in the session
	session SessionManager // The session manager used to manage session data
}

// NewFlashManager creates a new FlashManager instance.
func NewFlashManager(key string, session SessionManager) *FlashManager {
	return &FlashManager{key: key, session: session}
}

// Add adds a new flash message to the session.
func (fm *FlashManager) Add(ctx context.Context, msgType Type, content string) {
	var messages Messages
	if v := fm.session.Get(ctx, fm.key); v != nil {
		messages = v.(Messages)
	}

	messages = append(messages, Message{
		Type:    msgType,
		Content: content,
	})

	fm.session.Put(ctx, fm.key, messages)
}

// Get retrieves flash messages from the session without clearing them.
func (fm *FlashManager) Get(ctx context.Context) Messages {
	var messages Messages
	if v := fm.session.Get(ctx, fm.key); v != nil {
		messages = v.(Messages)
	}
	return messages
}

// Pop retrieves and clears flash messages from the session.
func (fm *FlashManager) Pop(ctx context.Context) Messages {
	var messages Messages
	if v := fm.session.Pop(ctx, fm.key); v != nil {
		messages = v.(Messages)
	}
	return messages
}

// Clear removes flash messages from the session.
func (fm *FlashManager) Clear(ctx context.Context) {
	fm.session.Put(ctx, fm.key, Messages{})
}

// AddAlert adds an alert flash message.
func (fm *FlashManager) AddAlert(ctx context.Context, content string) {
	fm.Add(ctx, TypeAlert, content)
}

// AddNotice adds a notice flash message.
func (fm *FlashManager) AddNotice(ctx context.Context, content string) {
	fm.Add(ctx, TypeNotice, content)
}
