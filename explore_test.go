package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"pokedex-go/internal/pokeapi"
	"pokedex-go/internal/pokecache"

	"pgregory.net/rapid"
)

// panicTransport is an http.RoundTripper that panics if any HTTP request is made,
// ensuring that no API call occurs during area validation rejection.
type panicTransport struct{}

func (panicTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(strings.NewReader("")),
	}, fmt.Errorf("unexpected API call made during area validation")
}

// TestProperty_AreaValidationRejects validates Property 5: Area Validation Rejects Invalid Areas.
// **Validates: Requirements 3.2**
// Feature: location-movement-system, Property 5: Area Validation Rejects Invalid Areas
func TestProperty_AreaValidationRejects(t *testing.T) {
	// Override HTTP transport to detect any unexpected API calls
	origTransport := http.DefaultTransport
	http.DefaultTransport = panicTransport{}
	defer func() { http.DefaultTransport = origTransport }()

	rapid.Check(t, func(t *rapid.T) {
		// Generate a random number of areas (1-50)
		numAreas := rapid.IntRange(1, 50).Draw(t, "numAreas")

		// Generate random area names for the location
		areas := make([]struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}, numAreas)
		areaNames := make(map[string]bool, numAreas)
		for i := range areas {
			name := rapid.StringMatching("[a-z][a-z0-9-]{2,20}").Draw(t, fmt.Sprintf("areaName-%d", i))
			areas[i].Name = name
			areas[i].URL = fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", name)
			areaNames[name] = true
		}

		// Generate a random location name
		locName := rapid.StringMatching("[a-z][a-z0-9-]{2,20}").Draw(t, "locName")

		loc := pokeapi.Location{
			ID:    1,
			Name:  locName,
			Areas: areas,
			Region: struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{Name: "test-region", URL: "https://pokeapi.co/api/v2/region/1"},
		}

		// Generate an invalid area name that is NOT in the location's areas list
		invalidArea := rapid.StringMatching("[a-z][a-z0-9-]{2,20}").Draw(t, "invalidArea")
		// Keep generating until we get one that's not in the areas list
		for areaNames[invalidArea] {
			invalidArea = rapid.StringMatching("[a-z][a-z0-9-]{2,20}").Draw(t, "invalidAreaRetry")
		}

		// Set up initial state
		initialArea := rapid.StringMatching("[a-z][a-z0-9-]{0,10}").Draw(t, "initialArea")

		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentLocation: &loc,
			CurrentArea:     initialArea,
			ScopedPageIndex: 0,
		}
		term := newMockTerminal()

		// Call commandExplore with the invalid area name
		err := commandExplore(cfg, []string{invalidArea}, term)
		if err != nil {
			t.Fatalf("commandExplore returned unexpected error: %v", err)
		}

		// Assert: output contains error message about the area not being part of the location
		combined := strings.Join(term.Calls, "")
		if !strings.Contains(combined, "is not part of") {
			t.Fatalf("expected 'is not part of' error message, got: %s", combined)
		}
		if !strings.Contains(combined, invalidArea) {
			t.Fatalf("expected error message to contain invalid area name %q, got: %s", invalidArea, combined)
		}
		if !strings.Contains(combined, locName) {
			t.Fatalf("expected error message to contain location name %q, got: %s", locName, combined)
		}

		// Assert: cfg.CurrentArea remains unchanged (no state mutation)
		if cfg.CurrentArea != initialArea {
			t.Fatalf("expected CurrentArea=%q to remain unchanged, got %q", initialArea, cfg.CurrentArea)
		}

		// Note: No API call was made because panicTransport would return an error
		// if GetLocationAreaDetails were called, and the test would fail at that point.
		// The fact that we reach here confirms no API call was attempted.
	})
}

// errorTransport is an http.RoundTripper that always returns a connection error,
// simulating an API failure for property testing.
type errorTransport struct{}

func (errorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("simulated network failure")
}

// TestProperty_FailedExplorePreservesState validates Property 7: Failed Explore Preserves State.
// **Validates: Requirements 3.5**
// Feature: location-movement-system, Property 7: Failed Explore Preserves State
func TestProperty_FailedExplorePreservesState(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random number of areas (1-50)
		numAreas := rapid.IntRange(1, 50).Draw(t, "numAreas")

		// Generate random area names for the location
		areas := make([]struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}, numAreas)
		for i := range areas {
			areas[i].Name = rapid.StringMatching("[a-z][a-z0-9-]{2,20}").Draw(t, fmt.Sprintf("areaName-%d", i))
			areas[i].URL = fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areas[i].Name)
		}

		// Generate a random location name
		locName := rapid.StringMatching("[a-z][a-z0-9-]{2,20}").Draw(t, "locName")

		loc := pokeapi.Location{
			ID:    1,
			Name:  locName,
			Areas: areas,
			Region: struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{Name: "test-region", URL: "https://pokeapi.co/api/v2/region/1"},
		}

		// Pick a VALID area from the list (so validation passes)
		chosenIdx := rapid.IntRange(0, numAreas-1).Draw(t, "chosenIdx")
		chosenArea := areas[chosenIdx].Name

		// Generate a random initial CurrentArea value
		initialArea := rapid.StringMatching("[a-z][a-z0-9-]{0,10}").Draw(t, "initialArea")

		// Use an empty cache so the area is NOT cached — forces an HTTP request
		testCache := pokecache.NewCache(60000)

		// Swap global cache
		oldCache := cache
		cache = testCache
		defer func() { cache = oldCache }()

		// Override HTTP transport to simulate API failure
		origTransport := http.DefaultTransport
		http.DefaultTransport = errorTransport{}
		defer func() { http.DefaultTransport = origTransport }()

		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentLocation: &loc,
			CurrentArea:     initialArea,
			ScopedPageIndex: 0,
		}
		term := newMockTerminal()

		// Call commandExplore with a valid area (validation passes, but API fails)
		err := commandExplore(cfg, []string{chosenArea}, term)
		if err != nil {
			t.Fatalf("commandExplore returned unexpected error: %v", err)
		}

		// Assert: cfg.CurrentArea remains unchanged
		if cfg.CurrentArea != initialArea {
			t.Fatalf("expected CurrentArea=%q to remain unchanged, got %q", initialArea, cfg.CurrentArea)
		}

		// Assert: output contains error message about fetching area details
		combined := strings.Join(term.Calls, "")
		if !strings.Contains(combined, "Error fetching area details") {
			t.Fatalf("expected error message containing 'Error fetching area details', got: %s", combined)
		}
	})
}

// TestProperty_ExploreUpdatesArea validates Property 6: Successful Explore Updates Active Area.
// **Validates: Requirements 3.1, 3.4**
// Feature: location-movement-system, Property 6: Successful Explore Updates Active Area
func TestProperty_ExploreUpdatesArea(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random number of areas (1-50)
		numAreas := rapid.IntRange(1, 50).Draw(t, "numAreas")

		// Generate random area names for the location
		areas := make([]struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}, numAreas)
		for i := range areas {
			areas[i].Name = rapid.StringMatching("[a-z][a-z0-9-]{2,20}").Draw(t, fmt.Sprintf("areaName-%d", i))
			areas[i].URL = fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areas[i].Name)
		}

		// Generate a random location name
		locName := rapid.StringMatching("[a-z][a-z0-9-]{2,20}").Draw(t, "locName")

		loc := pokeapi.Location{
			ID:    1,
			Name:  locName,
			Areas: areas,
			Region: struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{Name: "test-region", URL: "https://pokeapi.co/api/v2/region/1"},
		}

		// Pick a valid area at random from the list
		chosenIdx := rapid.IntRange(0, numAreas-1).Draw(t, "chosenIdx")
		chosenArea := areas[chosenIdx].Name

		// Create a minimal valid LocationAreaDetails for the chosen area
		areaDetails := pokeapi.LocationAreaDetails{
			ID:   1,
			Name: chosenArea,
			PokemonEncounters: []struct {
				Pokemon struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"pokemon"`
				VersionDetails []struct {
					Version struct {
						Name string `json:"name"`
						URL  string `json:"url"`
					} `json:"version"`
					MaxChance        int `json:"max_chance"`
					EncounterDetails []struct {
						MinLevel        int   `json:"min_level"`
						MaxLevel        int   `json:"max_level"`
						ConditionValues []any `json:"condition_values"`
						Chance          int   `json:"chance"`
						Method          struct {
							Name string `json:"name"`
							URL  string `json:"url"`
						} `json:"method"`
					} `json:"encounter_details"`
				} `json:"version_details"`
			}{
				{
					Pokemon: struct {
						Name string `json:"name"`
						URL  string `json:"url"`
					}{Name: "pikachu", URL: "https://pokeapi.co/api/v2/pokemon/25"},
				},
			},
		}

		jsonBytes, err := json.Marshal(areaDetails)
		if err != nil {
			t.Fatalf("failed to marshal area details: %v", err)
		}

		// Pre-populate the cache with the area details
		testCache := pokecache.NewCache(60000)
		url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", chosenArea)
		testCache.Add(url, jsonBytes)

		// Swap the global cache
		oldCache := cache
		cache = testCache
		defer func() { cache = oldCache }()

		// Set up Config with the location already set
		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentLocation: &loc,
			CurrentArea:     "",
			ScopedPageIndex: 0,
		}
		term := newMockTerminal()

		// Call commandExplore with the valid area name
		err = commandExplore(cfg, []string{chosenArea}, term)
		if err != nil {
			t.Fatalf("commandExplore returned error: %v", err)
		}

		// Assert: cfg.CurrentArea equals the chosen area
		if cfg.CurrentArea != chosenArea {
			t.Fatalf("expected CurrentArea=%q, got %q", chosenArea, cfg.CurrentArea)
		}

		// Assert: output contains "Pokemons in" confirming success path
		combined := strings.Join(term.Calls, "")
		if !strings.Contains(combined, "Pokemons in") {
			t.Fatalf("expected output to contain 'Pokemons in', got: %s", combined)
		}
	})
}
