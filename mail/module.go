package mail

import (
	"context"
)

type Module struct {
	config *Config
	mailer *Mailer
}

func NewMailerModule(config *Config) *Module {
	return &Module{
		config: config,
	}
}

func (m *Module) ID() string {
	return "hop.mail"
}

func (m *Module) Init() error {
	mailer, err := NewMailer(m.config)
	if err != nil {
		return err
	}
	m.mailer = mailer
	return nil
}

func (m *Module) Start(_ context.Context) error {
	return nil
}

func (m *Module) Stop(_ context.Context) error {
	return nil
}

func (m *Module) Mailer() *Mailer {
	return m.mailer
}
