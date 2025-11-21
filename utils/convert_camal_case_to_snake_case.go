package utils

import (
	"regexp"
	"strings"
)

func ConvertCamelCasetoSnakeCase(s string) string {
	re := regexp.MustCompile(`([A-Z])([A-Za-z0-9])`)

	word := re.ReplaceAllStringFunc(s, func(match string) string {
		return "_" + strings.ToLower(match[0:])
	})
	// to remove _staff_id -> staff_id
	return word[1:]
}
