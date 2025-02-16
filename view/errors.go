package view

type hyperViewError string

func (e hyperViewError) Error() string {
	return string(e)
}

const (
	// ErrTempNotFound is returned when a template is not found.
	ErrTempNotFound = hyperViewError("template not found")

	// ErrTempParse is returned when a template cannot be parsed.
	ErrTempParse = hyperViewError("template parse error")

	// ErrTempRender is returned when a template cannot be rendered.
	ErrTempRender = hyperViewError("template render error")
)
