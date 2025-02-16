package view

// Directory configuration that can be overridden
var (
	// LayoutsDir is the directory for layout templates
	LayoutsDir = "layouts"

	// PartialsDir is the directory for partial templates
	PartialsDir = "partials"

	// ViewsDir is the directory for view templates
	ViewsDir = "views"

	// SystemDir is the directory for system templates
	SystemDir = "system"

	// DefaultBaseLayout is the default base layout template
	DefaultBaseLayout = "base"
)

// Example of how to override the directories:
//
//
// import "minmarks.com/internal/view"
//
// func init() {
//     view.LayoutsDir = "custom/layouts"
//     view.PartialsDir = "custom/partials"
//     view.ViewsDir = "custom/views"
//     view.SystemDir = "custom/system"
//     view.DefaultBaseLayout = "custom_base"
// }

// Constants that should not be changed at runtime
const (
	// NonceContextKey is the key used for the front-end nonce
	NonceContextKey = "hop_nonce"
)
