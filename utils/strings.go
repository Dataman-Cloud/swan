package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// A domain name can only contain the letters A-Z, the digits 0-9 and hyphen (-).
const legalDomainChars = "ABCDEFGHIGKLMNOPQRSTUVWXYZabcdefghigklmnopqrstuvwxyz0123456789-"

func StripSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			// if the character is a space, drop it
			return -1
		}
		// else keep it in the string
		return r
	}, str)
}

func LegalDomain(str string) error {
	for i := 0; i < len(str); i++ {
		if !strings.Contains(legalDomainChars, string(str[i])) {
			return fmt.Errorf("character %q not allowed.", str[i])
		}
	}

	return nil
}

func RandomString(n int) string {
	b := make([]byte, 32)
	r := rand.Reader
	for {
		if _, err := io.ReadFull(r, b); err != nil {
			panic(err) // This shouldn't happen
		}

		id := hex.EncodeToString(b)

		if i := strings.IndexRune(id, ':'); i >= 0 {
			id = id[i+1:]
		}

		if len(id) > n {
			id = id[:n]
		}

		// if we try to parse the truncated for as an int and we don't have
		// an error then the value is all numeric and causes issues when
		// used as a hostname. ref #3869

		if _, err := strconv.ParseInt(id, 10, 64); err == nil {
			continue
		}

		return id
	}
}
