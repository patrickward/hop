package view

import "fmt"

type MessageType string

const (
	MessageError   MessageType = "error"
	MessageSuccess MessageType = "success"
	MessageWarning MessageType = "warning"
	MessageInfo    MessageType = "info"
)

// Message represents a message to be displayed to the user
type Message struct {
	Type    MessageType `json:"type"`
	Content string      `json:"content"`
}

// PageData provides a view model approach for passing data to templates
type PageData struct {
	title       string
	nonce       string
	description string
	data        map[string]any
	messages    []Message
	fieldErrors map[string]string
}

// NewPageData returns a new PageData struct
func NewPageData() *PageData {
	return &PageData{
		data:        make(map[string]any),
		fieldErrors: make(map[string]string),
	}
}

// Nonce returns the nonce value
func (p *PageData) Nonce() string {
	return p.nonce
}

// SetNonce sets the nonce value
func (p *PageData) SetNonce(nonce string) *PageData {
	p.nonce = nonce
	return p
}

// Title returns the title of the page.
func (p *PageData) Title() string {
	return p.title
}

// SetTitle sets the title of the page.
func (p *PageData) SetTitle(title string) *PageData {
	p.title = title
	return p
}

// Description returns the description of the page.
func (p *PageData) Description() string {
	return p.description
}

// SetDescription sets the description of the page.
func (p *PageData) SetDescription(desc string) *PageData {
	p.description = desc
	return p
}

// Data returns the data map
func (p *PageData) Data() map[string]any {
	return p.data
}

// SetData resets the data map with the provided data
func (p *PageData) SetData(data map[string]any) *PageData {
	p.data = data
	return p
}

// Set sets a key-value pair in the data map
func (p *PageData) Set(key string, value any) *PageData {
	p.data[key] = value
	return p
}

// Merge adds a map of data to the existing data map
func (p *PageData) Merge(data map[string]any) *PageData {
	for k, v := range data {
		p.data[k] = v
	}
	return p
}

// Get returns the value of the specified key from the data map
func (p *PageData) Get(key string) any {
	return p.data[key]
}

// Messages returns the messages slice
func (p *PageData) Messages() []Message {
	return p.messages
}

// SetError adds an error message to the messages slice
func (p *PageData) SetError(content string) *PageData {
	p.messages = append(p.messages, Message{
		Type:    MessageError,
		Content: content,
	})
	return p
}

// SetSuccess adds a success message to the messages slice
func (p *PageData) SetSuccess(content string) *PageData {
	p.messages = append(p.messages, Message{
		Type:    MessageSuccess,
		Content: content,
	})
	return p
}

// SetWarning adds a warning message to the messages slice
func (p *PageData) SetWarning(content string) *PageData {
	p.messages = append(p.messages, Message{
		Type:    MessageWarning,
		Content: content,
	})
	return p
}

// SetInfo adds an info message to the messages slice
func (p *PageData) SetInfo(content string) *PageData {
	p.messages = append(p.messages, Message{
		Type:    MessageInfo,
		Content: content,
	})
	return p
}

// HasMessages returns true if there are messages in the messages slice
func (p *PageData) HasMessages() bool {
	return len(p.messages) > 0
}

// AddFieldErrors adds a map of field errors to the messages slice. The map should be in the format of field name to error message.
func (p *PageData) AddFieldErrors(errors map[string]string) *PageData {
	for field, msg := range errors {
		p.fieldErrors[field] = msg
	}
	return p
}

// AddFieldError adds a field error to the messages slice
func (p *PageData) AddFieldError(field, msg string) *PageData {
	p.fieldErrors[field] = msg
	return p
}

// HasFieldErrors returns true if there are field errors in the messages slice
func (p *PageData) HasFieldErrors() bool {
	return len(p.fieldErrors) > 0
}

// HasErrorFor returns true if there is an error message for the specified field
func (p *PageData) HasErrorFor(field string) bool {
	_, ok := p.fieldErrors[field]
	return ok
}

// ErrorFor returns the error message for the specified field
func (p *PageData) ErrorFor(field string) string {
	return p.fieldErrors[field]
}

// FieldErrors returns a map of field errors
func (p *PageData) FieldErrors() map[string]string {
	return p.fieldErrors
}

// MessagesOfType returns messages of the specified type
func (p *PageData) MessagesOfType(t MessageType) []Message {
	messages := make([]Message, 0)
	for _, msg := range p.messages {
		if msg.Type == t {
			messages = append(messages, msg)
		}
	}
	return messages
}

// Clear resets the PageData struct
func (p *PageData) Clear() *PageData {
	p.title = ""
	p.description = ""
	p.data = make(map[string]any)
	p.messages = make([]Message, 0)
	p.fieldErrors = make(map[string]string)
	return p
}

// ClearMessages removes all messages from the messages slice
func (p *PageData) ClearMessages() *PageData {
	p.messages = make([]Message, 0)
	return p
}

// ClearFieldErrors removes all field errors from the messages slice
func (p *PageData) ClearFieldErrors() *PageData {
	p.fieldErrors = make(map[string]string)
	return p
}

// HxNonce returns an HTMX nonce value, if available.
// This adds the inlineScriptNonce key to a JSON object with the nonce value and can be used in an HTMX meta tag.
func (p *PageData) HxNonce() string {
	return fmt.Sprintf("{\"includeIndicatorStyles\":false,\"inlineScriptNonce\": \"%s\"}", p.Nonce())
}
