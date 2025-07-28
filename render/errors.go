package render

import "errors"

var (
	// ErrTemplateNotFound is returned when a template is not found.
	ErrTemplateNotFound = errors.New("template not found")

	// ErrTemplateNotParsed is returned when a template cannot be parsed.
	ErrTemplateNotParsed = errors.New("template not parsed")
)
