package main

import (
	"testing"

	"pgregory.net/rapid"
)

// TestProperty_NameTruncation verifies the truncateName function behavior.
// **Validates: Requirements 4.6**
func TestProperty_NameTruncation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		name := rapid.String().Draw(t, "name")
		result := truncateName(name, MaxNameLen)
		if len(name) > MaxNameLen {
			if len(result) != MaxNameLen {
				t.Fatalf("expected length %d, got %d", MaxNameLen, len(result))
			}
			if result[MaxNameLen-3:] != "..." {
				t.Fatalf("expected suffix '...', got '%s'", result[MaxNameLen-3:])
			}
		} else {
			if result != name {
				t.Fatalf("expected unchanged string %q, got %q", name, result)
			}
		}
	})
}
