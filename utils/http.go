package utils

import (
	"net/http"
	"strings"
)

func SameSiteFromString(key string) http.SameSite {
	switch strings.ToLower(key) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
