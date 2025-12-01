package user

import (
	"strings"
	"unicode"
)

// InitialsFromEmail generates initials from the email address
func InitialsFromEmail(email string) string {
	// extract local part
	at := strings.Index(email, "@")
	if at == -1 {
		return ""
	}
	local := email[:at]

	// normalize separators
	// replace . _ - with spaces to simplify splitting
	replacer := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	local = replacer.Replace(local)

	parts := strings.Fields(local)
	if len(parts) == 0 {
		return ""
	}

	// first and last non-empty parts
	first := parts[0]
	last := parts[len(parts)-1]

	// extract runes in case unicode shows up
	firstRune := []rune(first)
	lastRune := []rune(last)

	if len(firstRune) == 0 || len(lastRune) == 0 {
		return ""
	}

	// uppercase initials
	return strings.ToUpper(string(firstRune[0])) + strings.ToUpper(string(lastRune[0]))
}

// NameFromEmail tries to extract first and last name from the email address
func NameFromEmail(email string) string {
	at := strings.Index(email, "@")
	if at == -1 {
		return ""
	}

	local := email[:at]
	domain := email[at+1:]
	domain = strings.Split(domain, ".")[0] // just root domain

	// normalize separators
	replacer := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	local = replacer.Replace(local)

	parts := strings.Fields(local)
	if len(parts) == 0 {
		return ""
	}

	// if there are two or more meaningful username parts → first + last
	if len(parts) >= 2 {
		return cleanWord(parts[0]) + " " + cleanWord(parts[len(parts)-1])
	}

	// only one word in the username
	one := cleanWord(parts[0])

	// local == domain → probably a name
	if strings.EqualFold(one, domain) {
		return title(one)
	}

	// username is a single letter: "m"
	if len([]rune(one)) == 1 {
		return strings.ToUpper(one)
	}

	// username looks like "zack"
	return title(one)
}

func cleanWord(s string) string {
	// strip trailing digits
	for len(s) > 0 && unicode.IsDigit(rune(s[len(s)-1])) {
		s = s[:len(s)-1]
	}
	return title(s)
}

func title(s string) string {
	r := []rune(strings.ToLower(s))
	if len(r) == 0 {
		return ""
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
