package sanitize

import (
	"html"
	"strings"
	"unicode/utf8"
)

const (
	MaxNameLength  = 64
	MaxUsernameLen = 32
	MaxTextLength  = 4096
)

func String(s string) string {
	s = strings.TrimSpace(s)
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "")
	}
	s = html.EscapeString(s)
	if len([]rune(s)) > MaxTextLength {
		s = string([]rune(s)[:MaxTextLength])
	}
	return s
}

func Name(s string) string {
	s = strings.TrimSpace(s)
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "")
	}
	s = html.EscapeString(s)
	if len([]rune(s)) > MaxNameLength {
		s = string([]rune(s)[:MaxNameLength])
	}
	return s
}

func Username(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "@")
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "")
	}
	s = html.EscapeString(s)
	if len([]rune(s)) > MaxUsernameLen {
		s = string([]rune(s)[:MaxUsernameLen])
	}
	return s
}

func CallbackData(data string) bool {
	if len(data) == 0 {
		return false
	}
	parts := strings.Split(data, "_")
	if len(parts) != 2 {
		return false
	}
	action := parts[0]
	allowed := map[string]bool{
		"approve": true, "reject": true,
		"makeadmin": true, "removeadmin": true,
		"block": true, "unblock": true, "delete": true,
	}
	if !allowed[action] {
		return false
	}
	idStr := parts[1]
	for _, c := range idStr {
		if c < '0' || c > '9' {
			return false
		}
	}
	if len(idStr) == 0 || len(idStr) > 15 {
		return false
	}
	return true
}
