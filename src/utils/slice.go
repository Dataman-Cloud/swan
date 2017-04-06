package utils

import (
	"strings"
)

func SliceUnique(slice []string) bool {
	m := make(map[string]int)
	for _, s := range slice {
		if _, ok := m[s]; ok {
			return true
		}
		m[s] = 1
	}
	return false
}

func SliceContains(slice []string, value string) bool {
	for _, s := range slice {
		if strings.TrimSpace(s) == strings.TrimSpace(value) {
			return true
		}
	}

	return false
}
