package utils

import (
	"fmt"
	"strings"
)

func EscapeString(s string) string {
	m := map[string]string{
		"_": "\\_",
		"*": "\\*",
		"[": "\\[",
		"]": "\\]",
		"(": "\\(",
		")": "\\)",
		"~": "\\~",
		"`": "\\`",
		">": "\\>",
		"#": "\\#",
		"+": "\\+",
		"-": "\\-",
		"=": "\\=",
		"|": "\\|",
		"{": "\\{",
		"}": "\\}",
		".": "\\.",
		"!": "\\!",
	}
	for k, v := range m {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}

func ParseCommand(s string) (string, []string, error) {
	if s[0] != '/' {
		return "", nil, fmt.Errorf("command %s should be started with '/': %s", s, s)
	}
	arr := strings.SplitN(s, " ", 2)
	if len(arr) < 2 {
		return arr[0], nil, fmt.Errorf("command %s should have arguments splitted by comma", s)
	}
	return arr[0], strings.Split(arr[1], ", "), nil
}

func MapToSlice[K comparable, V any](m map[K]V) []V {
	s := make([]V, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}
