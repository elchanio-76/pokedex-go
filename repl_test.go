package main

import (
	_ "fmt"
	"testing"
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