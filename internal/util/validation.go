package util

import (
	"regexp"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func IsValidUUID(s string) bool {
	if s == "" {
		return false
	}
	return uuidRegex.MatchString(s)
}

func IsValidEnum(value string, validValues []string) bool {
	if value == "" {
		return true
	}
	for _, v := range validValues {
		if value == v {
			return true
		}
	}
	return false
}
