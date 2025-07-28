package alert_test

import (
	"context"
	"testing"

	"github.com/patrickward/hop/v2/alert"
)

func TestFlash(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	flash := alert.NewFlashManager("flash", &sessionManagerMock{})

	// Add a message
	flash.Add(ctx, alert.TypeSuccess, "Operation completed successfully")

	// Check if the message was added
	messages := flash.Get(ctx)

	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Type != alert.TypeSuccess || messages[0].Content != "Operation completed successfully" {
		t.Errorf("Expected message type %s with content 'Operation completed successfully', got type %s with content '%s'",
			alert.TypeSuccess, messages[0].Type, messages[0].Content)
	}

	// Pop the message
	poppedMessages := flash.Pop(ctx)
	if len(poppedMessages) != 1 {
		t.Errorf("Expected 1 popped message, got %d", len(poppedMessages))
	}

	if poppedMessages[0].Type != alert.TypeSuccess || poppedMessages[0].Content != "Operation completed successfully" {
		t.Errorf("Expected popped message type %s with content 'Operation completed successfully', got type %s with content '%s'",
			alert.TypeSuccess, poppedMessages[0].Type, poppedMessages[0].Content)
	}
}

type sessionManagerMock struct {
	sessionData map[string]interface{}
}

func (s *sessionManagerMock) Get(_ context.Context, key string) interface{} {
	if val, exists := s.sessionData[key]; exists {
		return val
	}
	return nil
}

func (s *sessionManagerMock) Pop(_ context.Context, key string) interface{} {
	if val, exists := s.sessionData[key]; exists {
		delete(s.sessionData, key)
		return val
	}
	return nil
}

func (s *sessionManagerMock) Put(_ context.Context, key string, val interface{}) {
	if s.sessionData == nil {
		s.sessionData = make(map[string]interface{})
	}
	s.sessionData[key] = val
}
