package pipeline_test

import (
	"encoding/json"
	"strings"
)

func decodeJSON(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}

func containsAny(text string, needles ...string) bool {
	text = strings.ToLower(text)
	for _, n := range needles {
		if strings.Contains(text, strings.ToLower(n)) {
			return true
		}
	}
	return false
}
