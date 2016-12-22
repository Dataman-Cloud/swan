package utils

import (
	"strings"
)

func SliceUnique(slice []string) bool {
	m := make(map[string]int)

	sliceLen := len(slice)

	for _, s := range slice {
		m[s] = 1
	}

	return sliceLen == len(m)
}

func SliceContains(slice []string, value string) bool {
	for _, s := range slice {
		if strings.TrimSpace(s) == strings.TrimSpace(value) {
			return true
		}
	}

	return false
}
