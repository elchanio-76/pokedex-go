package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"pokedex-go/internal/input"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		name		 string
		input    string
		expected []string
	}{
		{
			name:			"Words with extra spaces",
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			name:		"Capitalized input",
			input:	"Charmander Bulbasaur PIKACHU",
			expected: []string{"charmander", "bulbasaur", "pikachu"},
		},
		{
			name:		"Empty input",
			input:	"",
			expected: []string{},
		},
		{
			name:		"Special characters",
			input:	"--Charmander + Bulbasaur ! PIKACHU",
			expected: []string{"--charmander", "+", "bulbasaur", "!", "pikachu"},
		},
	}
	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("FAIL: %s - Expected %d, got %d items\n", c.name, len(c.expected), len(actual))
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if expectedWord != word {
				t.Errorf("FAIL: %s - Expected %s, got %s\n", c.name, expectedWord, word)
			}
		}
	}

}

// mockTerminal implements input.Terminal for testing.
// ReadLine returns lines from the provided slice, then io.EOF.
// Print and Println append formatted output to Calls.
type mockTerminal struct {
	lines  []string
	cursor int
	Calls  []string
}

func newMockTerminal(lines ...string) *mockTerminal {
	return &mockTerminal{lines: lines}
}

func (m *mockTerminal) ReadLine(prompt string) (string, error) {
	if m.cursor >= len(m.lines) {
		return "", io.EOF
	}
	line := m.lines[m.cursor]
	m.cursor++
	return line, nil
}

func (m *mockTerminal) Print(format string, args ...any) {
	m.Calls = append(m.Calls, fmt.Sprintf(format, args...))
}

func (m *mockTerminal) Println(args ...any) {
	m.Calls = append(m.Calls, fmt.Sprintln(args...))
}

func (m *mockTerminal) Close() error        { return nil }
func (m *mockTerminal) SetHistory([]string) {}

// compile-time check
var _ input.Terminal = (*mockTerminal)(nil)

// TestREPL_UnknownCommand verifies that an unrecognised command causes
// startREPL to call Print with an "Unknown command" message (Req 4.1).
func TestREPL_UnknownCommand(t *testing.T) {
	term := newMockTerminal("notacommand")
	cfg := &Config{
		Next:    "",
		Prev:    "",
		Pokedex: make(map[string]string),
	}

	startREPL(cfg, term)

	found := false
	for _, call := range term.Calls {
		if strings.Contains(call, "Unknown command") && strings.Contains(call, "notacommand") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Unknown command notacommand' in output, got: %v", term.Calls)
	}
}

// TestREPL_CallbackError verifies that when a callback returns an error,
// startREPL calls Println with that error (Req 4.2).
func TestREPL_CallbackError(t *testing.T) {
	// Register a temporary command that always errors.
	sentinel := errors.New("boom")
	commandRegistry["_testerr"] = cliCommand{
		name:        "_testerr",
		description: "test error command",
		callback: func(cfg *Config, args []string, t input.Terminal) error {
			return sentinel
		},
	}
	defer delete(commandRegistry, "_testerr")

	term := newMockTerminal("_testerr")
	cfg := &Config{Pokedex: make(map[string]string)}

	startREPL(cfg, term)

	found := false
	for _, call := range term.Calls {
		if strings.Contains(call, "boom") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected error 'boom' in output, got: %v", term.Calls)
	}
}

// TestCommandHelp_UsesPrint verifies that commandHelp calls Print with
// help text (Req 5.2).
func TestCommandHelp_UsesPrint(t *testing.T) {
	term := newMockTerminal()
	cfg := &Config{Pokedex: make(map[string]string)}

	err := commandHelp(cfg, nil, term)
	if err != nil {
		t.Fatalf("commandHelp returned unexpected error: %v", err)
	}

	combined := strings.Join(term.Calls, "")
	if !strings.Contains(combined, "Welcome to the Pokedex") {
		t.Errorf("expected help text in output, got: %v", term.Calls)
	}
	if len(term.Calls) == 0 {
		t.Error("expected at least one Print call from commandHelp")
	}
}
