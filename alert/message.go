package alert

type Message struct {
	Type    Type   `json:"type"`    // Type of message (e.g., "info", "error", "success")
	Content string `json:"content"` // Content of the message
}

type Type string

const (
	TypeError   Type = "error"
	TypeSuccess Type = "success"
	TypeWarning Type = "warning"
	TypeInfo    Type = "info"
	TypeAlert   Type = "alert"
	TypeNotice  Type = "notice"
)

// Messages is a slice of Message.
type Messages []Message

// ByType returns messages filtered by type.
func (m Messages) ByType(t Type) Messages {
	var filtered Messages
	for _, msg := range m {
		if msg.Type == t {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}
