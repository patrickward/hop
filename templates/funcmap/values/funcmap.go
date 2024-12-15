package values

import "html/template"

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"val_yesno":    yesno,
		"val_onoff":    onoff,
		"val_coalesce": coalesce,
	}
}

// coalesce returns the first non-empty value
func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

// yesno returns "Yes" or "No" based on the boolean value
func yesno(b bool) string {
	if b {
		return "Yes"
	}

	return "No"
}

// onoff returns "On" or "Off" based on the boolean value
func onoff(b bool) string {
	if b {
		return "On"
	}

	return "Off"
}
