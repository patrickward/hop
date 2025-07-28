package render_test

import (
	"context"
)

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
