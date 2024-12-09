package mail

import (
	"fmt"
)

// TemplateError provides context about template errors
type TemplateError struct {
	TemplateName string
	OriginalErr  error
	Phase        string // "parse", "execute", "process"
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf("template error in %s during %s phase: %v", e.TemplateName, e.Phase, e.OriginalErr)
}
