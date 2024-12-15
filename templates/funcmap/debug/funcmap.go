package debug

import (
	"encoding/json"
	"fmt"
	"html/template"
)

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"dbg_dump":   dump,   // pretty prints any value
		"dbg_typeof": typeof, // returns the type of a value
	}
}

// dump pretty prints any value
func dump(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

// typeof returns the type of a value
func typeof(v any) string {
	return fmt.Sprintf("%T", v)
}
