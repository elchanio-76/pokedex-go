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

// TestProperty_VisitStateTransition validates Property 1: Visit State Transition.
// **Validates: Requirements 1.1, 1.5, 2.8**
func TestProperty_VisitStateTransition(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		locName := rapid.StringMatching("[a-z][a-z0-9-]{0,29}").Draw(t, "locName")
		regionName := rapid.StringMatching("[a-z][a-z0-9-]{0,29}").Draw(t, "regionName")
		numAreas := rapid.IntRange(0, 50).Draw(t, "numAreas")

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
			Name:  locName,
			Areas: areas,
			Region: struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{Name: regionName, URL: "https://pokeapi.co/api/v2/region/1"},
		}

		jsonBytes, err := json.Marshal(loc)
		if err != nil {
			t.Fatalf("failed to marshal location: %v", err)
		}

		testCache := pokecache.NewCache(60000)
		url := fmt.Sprintf("https://pokeapi.co/api/v2/location/%s", locName)
		testCache.Add(url, jsonBytes)

		// Temporarily swap the global cache
		oldCache := cache
		cache = testCache
		defer func() { cache = oldCache }()

		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentArea:     "old-area",
			ScopedPageIndex: 5,
		}
		term := newMockTerminal()

		err = commandVisit(cfg, []string{locName}, term)
		if err != nil {
			t.Fatalf("commandVisit returned error: %v", err)
		}

		// Assert state transition
		if cfg.CurrentLocation == nil {
			t.Fatal("CurrentLocation should not be nil after successful visit")
		}
		if cfg.CurrentLocation.Name != locName {
			t.Fatalf("expected CurrentLocation.Name=%q, got %q", locName, cfg.CurrentLocation.Name)
		}
		if cfg.CurrentArea != "" {
			t.Fatalf("expected CurrentArea to be empty, got %q", cfg.CurrentArea)
		}
		if cfg.ScopedPageIndex != 0 {
			t.Fatalf("expected ScopedPageIndex=0, got %d", cfg.ScopedPageIndex)
		}
	})
}

// TestProperty_VisitOutputContainsRequiredInfo validates Property 2: Visit Output Contains Required Information.
// **Validates: Requirements 1.2**
func TestProperty_VisitOutputContainsRequiredInfo(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		locName := rapid.StringMatching("[a-z][a-z0-9-]{0,29}").Draw(t, "locName")
		regionName := rapid.StringMatching("[a-z][a-z0-9-]{0,29}").Draw(t, "regionName")
		numAreas := rapid.IntRange(0, 50).Draw(t, "numAreas")

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
			Name:  locName,
			Areas: areas,
			Region: struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{Name: regionName, URL: "https://pokeapi.co/api/v2/region/1"},
		}

		jsonBytes, err := json.Marshal(loc)
		if err != nil {
			t.Fatalf("failed to marshal location: %v", err)
		}

		testCache := pokecache.NewCache(60000)
		url := fmt.Sprintf("https://pokeapi.co/api/v2/location/%s", locName)
		testCache.Add(url, jsonBytes)

		oldCache := cache
		cache = testCache
		defer func() { cache = oldCache }()

		cfg := &Config{
			Pokedex: make(map[string]string),
		}
		term := newMockTerminal()

		err = commandVisit(cfg, []string{locName}, term)
		if err != nil {
			t.Fatalf("commandVisit returned error: %v", err)
		}

		combined := strings.Join(term.Calls, "")

		// Output must contain the location name
		if !strings.Contains(combined, locName) {
			t.Fatalf("output does not contain location name %q, got: %s", locName, combined)
		}

		// Output must contain the region name
		if !strings.Contains(combined, regionName) {
			t.Fatalf("output does not contain region name %q, got: %s", regionName, combined)
		}

		// Output must contain the area count
		areaCount := fmt.Sprintf("%d", numAreas)
		if !strings.Contains(combined, areaCount) {
			t.Fatalf("output does not contain area count %q, got: %s", areaCount, combined)
		}
	})
}

// failTransport is an http.RoundTripper that always returns a 404 response,
// allowing tests to simulate API failures without real network access.
type failTransport struct{}

func (failTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

// TestProperty_FailedVisitPreservesState validates Property 3: Failed Visit Preserves State.
// **Validates: Requirements 1.3**
func TestProperty_FailedVisitPreservesState(t *testing.T) {
	// Override HTTP transport to return 404 without network access
	origTransport := http.DefaultTransport
	http.DefaultTransport = failTransport{}
	defer func() { http.DefaultTransport = origTransport }()

	rapid.Check(t, func(t *rapid.T) {
		// Generate random initial state
		initialArea := rapid.StringMatching("[a-z][a-z0-9-]{0,19}").Draw(t, "initialArea")
		initialPageIndex := rapid.IntRange(0, 10).Draw(t, "initialPageIndex")

		// Create a location that may or may not be set initially
		hasLocation := rapid.Bool().Draw(t, "hasLocation")

		cfg := &Config{
			Pokedex:         make(map[string]string),
			CurrentArea:     initialArea,
			ScopedPageIndex: initialPageIndex,
		}

		if hasLocation {
			cfg.CurrentLocation = &pokeapi.Location{
				Name: "existing-loc",
			}
		}

		// Use an empty cache so GetLocation won't find the URL and will hit HTTP (which returns 404)
		testCache := pokecache.NewCache(60000)
		oldCache := cache
		cache = testCache
		defer func() { cache = oldCache }()

		term := newMockTerminal()

		// Use a random nonexistent location name that won't be in cache
		badName := rapid.StringMatching("[a-z]{4,10}-[a-z]{4,10}-nonexist").Draw(t, "badName")
		_ = commandVisit(cfg, []string{badName}, term)

		// Assert all Config fields remain unchanged
		if cfg.CurrentArea != initialArea {
			t.Fatalf("expected CurrentArea=%q, got %q", initialArea, cfg.CurrentArea)
		}
		if cfg.ScopedPageIndex != initialPageIndex {
			t.Fatalf("expected ScopedPageIndex=%d, got %d", initialPageIndex, cfg.ScopedPageIndex)
		}
		if hasLocation {
			if cfg.CurrentLocation == nil || cfg.CurrentLocation.Name != "existing-loc" {
				t.Fatalf("expected CurrentLocation to be preserved")
			}
		} else {
			if cfg.CurrentLocation != nil {
				t.Fatalf("expected CurrentLocation to remain nil")
			}
		}

		// Verify the error message was printed
		combined := strings.Join(term.Calls, "")
		if !strings.Contains(combined, "not found") {
			t.Fatalf("expected 'not found' error message, got: %s", combined)
		}
	})
}
