package main

import (
	"fmt"
	"strings"
	"testing"

	"pokedex-go/internal/pokeapi"

	"pgregory.net/rapid"
)

// TestProperty_PaginationCorrectness validates Property 4: Pagination Correctness.
// **Validates: Requirements 2.1, 2.2, 2.6**
func TestProperty_PaginationCorrectness(t *testing.T) {
	t.Run("commandMap returns correct slice and never exceeds 20 items", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			numAreas := rapid.IntRange(1, 100).Draw(t, "numAreas")
			maxPage := (numAreas-1)/20 + 1 // total number of pages
			pageIndex := rapid.IntRange(0, maxPage-1).Draw(t, "pageIndex")

			areas := make([]struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}, numAreas)
			for i := range areas {
				areas[i].Name = fmt.Sprintf("area-%d", i)
				areas[i].URL = fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%d", i)
			}

			loc := pokeapi.Location{
				ID:    1,
				Name:  "test-location",
				Areas: areas,
				Region: struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				}{Name: "test-region", URL: "https://pokeapi.co/api/v2/region/1"},
			}

			cfg := &Config{
				Pokedex:         make(map[string]string),
				CurrentLocation: &loc,
				ScopedPageIndex: pageIndex,
			}
			term := newMockTerminal()

			err := commandMap(cfg, nil, term)
			if err != nil {
				t.Fatalf("commandMap returned error: %v", err)
			}

			// Calculate expected slice
			start := pageIndex * 20
			end := (pageIndex + 1) * 20
			if end > numAreas {
				end = numAreas
			}
			expectedCount := end - start

			// Verify output never exceeds 20 items
			if len(term.Calls) > 20 {
				t.Fatalf("output has %d items, expected at most 20", len(term.Calls))
			}

			// Verify output matches expected count
			if len(term.Calls) != expectedCount {
				t.Fatalf("expected %d output lines, got %d", expectedCount, len(term.Calls))
			}

			// Verify each output line contains the correct area name
			for i, call := range term.Calls {
				expectedName := fmt.Sprintf("area-%d", start+i)
				if !strings.Contains(call, expectedName) {
					t.Fatalf("output line %d: expected to contain %q, got %q", i, expectedName, call)
				}
			}

			// Verify ScopedPageIndex incremented
			if cfg.ScopedPageIndex != pageIndex+1 {
				t.Fatalf("expected ScopedPageIndex=%d, got %d", pageIndex+1, cfg.ScopedPageIndex)
			}
		})
	})

	t.Run("commandMapBack returns correct slice and never exceeds 20 items", func(t *testing.T) {
		rapid.Check(t, func(t *rapid.T) {
			numAreas := rapid.IntRange(1, 100).Draw(t, "numAreas")
			maxPage := (numAreas-1)/20 + 1
			// pageIndex > 0 so mapb can go back
			pageIndex := rapid.IntRange(1, maxPage).Draw(t, "pageIndex")

			areas := make([]struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}, numAreas)
			for i := range areas {
				areas[i].Name = fmt.Sprintf("area-%d", i)
				areas[i].URL = fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%d", i)
			}

			loc := pokeapi.Location{
				ID:    1,
				Name:  "test-location",
				Areas: areas,
				Region: struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				}{Name: "test-region", URL: "https://pokeapi.co/api/v2/region/1"},
			}

			cfg := &Config{
				Pokedex:         make(map[string]string),
				CurrentLocation: &loc,
				ScopedPageIndex: pageIndex,
			}
			term := newMockTerminal()

			err := commandMapBack(cfg, nil, term)
			if err != nil {
				t.Fatalf("commandMapBack returned error: %v", err)
			}

			// After mapb, ScopedPageIndex decrements first, then displays that page
			expectedPageIndex := pageIndex - 1
			start := expectedPageIndex * 20
			end := (expectedPageIndex + 1) * 20
			if end > numAreas {
				end = numAreas
			}
			expectedCount := end - start

			// Verify output never exceeds 20 items
			if len(term.Calls) > 20 {
				t.Fatalf("output has %d items, expected at most 20", len(term.Calls))
			}

			// Verify output matches expected count
			if len(term.Calls) != expectedCount {
				t.Fatalf("expected %d output lines, got %d", expectedCount, len(term.Calls))
			}

			// Verify each output line contains the correct area name
			for i, call := range term.Calls {
				expectedName := fmt.Sprintf("area-%d", start+i)
				if !strings.Contains(call, expectedName) {
					t.Fatalf("output line %d: expected to contain %q, got %q", i, expectedName, call)
				}
			}

			// Verify ScopedPageIndex decremented
			if cfg.ScopedPageIndex != expectedPageIndex {
				t.Fatalf("expected ScopedPageIndex=%d, got %d", expectedPageIndex, cfg.ScopedPageIndex)
			}
		})
	})

	t.Run("commandMap with nil location prints error", func(t *testing.T) {
		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentLocation: nil,
		}
		term := newMockTerminal()

		err := commandMap(cfg, nil, term)
		if err != nil {
			t.Fatalf("commandMap returned error: %v", err)
		}

		combined := strings.Join(term.Calls, "")
		if !strings.Contains(combined, "No location set") {
			t.Fatalf("expected 'No location set' message, got: %s", combined)
		}
	})

	t.Run("commandMapBack with nil location prints error", func(t *testing.T) {
		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentLocation: nil,
		}
		term := newMockTerminal()

		err := commandMapBack(cfg, nil, term)
		if err != nil {
			t.Fatalf("commandMapBack returned error: %v", err)
		}

		combined := strings.Join(term.Calls, "")
		if !strings.Contains(combined, "No location set") {
			t.Fatalf("expected 'No location set' message, got: %s", combined)
		}
	})

	t.Run("commandMap with empty areas prints error", func(t *testing.T) {
		loc := pokeapi.Location{
			ID:   1,
			Name: "empty-location",
			Areas: []struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{},
		}
		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentLocation: &loc,
		}
		term := newMockTerminal()

		err := commandMap(cfg, nil, term)
		if err != nil {
			t.Fatalf("commandMap returned error: %v", err)
		}

		combined := strings.Join(term.Calls, "")
		if !strings.Contains(combined, "no areas") {
			t.Fatalf("expected 'no areas' message, got: %s", combined)
		}
	})

	t.Run("commandMapBack with empty areas prints error", func(t *testing.T) {
		loc := pokeapi.Location{
			ID:   1,
			Name: "empty-location",
			Areas: []struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{},
		}
		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentLocation: &loc,
		}
		term := newMockTerminal()

		err := commandMapBack(cfg, nil, term)
		if err != nil {
			t.Fatalf("commandMapBack returned error: %v", err)
		}

		combined := strings.Join(term.Calls, "")
		if !strings.Contains(combined, "no areas") {
			t.Fatalf("expected 'no areas' message, got: %s", combined)
		}
	})

	t.Run("commandMap on last page prints message", func(t *testing.T) {
		areas := make([]struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}, 5)
		for i := range areas {
			areas[i].Name = fmt.Sprintf("area-%d", i)
			areas[i].URL = fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%d", i)
		}

		loc := pokeapi.Location{
			ID:    1,
			Name:  "small-location",
			Areas: areas,
		}
		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentLocation: &loc,
			ScopedPageIndex: 1, // start >= len(areas) since 1*20=20 > 5
		}
		term := newMockTerminal()

		err := commandMap(cfg, nil, term)
		if err != nil {
			t.Fatalf("commandMap returned error: %v", err)
		}

		combined := strings.Join(term.Calls, "")
		if !strings.Contains(combined, "last page") {
			t.Fatalf("expected 'last page' message, got: %s", combined)
		}
	})

	t.Run("commandMapBack on first page prints message", func(t *testing.T) {
		areas := make([]struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}, 5)
		for i := range areas {
			areas[i].Name = fmt.Sprintf("area-%d", i)
			areas[i].URL = fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%d", i)
		}

		loc := pokeapi.Location{
			ID:    1,
			Name:  "small-location",
			Areas: areas,
		}
		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentLocation: &loc,
			ScopedPageIndex: 0,
		}
		term := newMockTerminal()

		err := commandMapBack(cfg, nil, term)
		if err != nil {
			t.Fatalf("commandMapBack returned error: %v", err)
		}

		combined := strings.Join(term.Calls, "")
		if !strings.Contains(combined, "first page") {
			t.Fatalf("expected 'first page' message, got: %s", combined)
		}
	})
}
