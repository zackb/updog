package id

import (
	"testing"
)

func TestNewID(t *testing.T) {
	id := NewID()
	if len(id) != 11 {
		t.Errorf("Expected ID length of 11, got %d", len(id))
	}

	for _, char := range id {
		if !isValidChar(char) {
			t.Errorf("Invalid character '%c' in ID", char)
		}
	}
}

func isValidChar(char rune) bool {
	validChars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for _, validChar := range validChars {
		if char == validChar {
			return true
		}
	}
	return false
}
