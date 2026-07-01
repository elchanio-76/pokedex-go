package pokeapi

import (
	"encoding/json"
	"fmt"
	"testing"

	"pokedex-go/internal/pokecache"

	"pgregory.net/rapid"
)

// TestProperty_GetLocationCaching verifies that GetLocation uses the cache correctly.
// Property 1: Visit State Transition (partial — cache storage)
// **Validates: Requirements 5.1, 5.2**
//
// For any random location name, when the cache already contains valid JSON for the
// corresponding URL key, GetLocation returns the cached Location without making an HTTP request.
func TestProperty_GetLocationCaching(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random location name (alphanumeric + hyphens, like real PokeAPI names)
		locationName := rapid.StringMatching(`[a-z][a-z0-9\-]{0,29}`).Draw(t, "locationName")

		// Generate random Location data to store in cache
		locationID := rapid.IntRange(1, 10000).Draw(t, "locationID")
		regionName := rapid.StringMatching(`[a-z][a-z\-]{2,15}`).Draw(t, "regionName")
		areaCount := rapid.IntRange(0, 10).Draw(t, "areaCount")

		areas := make([]struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		}, areaCount)
		for i := range areas {
			areas[i].Name = fmt.Sprintf("area-%d", i)
			areas[i].URL = fmt.Sprintf("https://pokeapi.co/api/v2/location-area/area-%d", i)
		}

		expectedLocation := Location{
			ID:    locationID,
			Name:  locationName,
			Areas: areas,
			Region: struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{
				Name: regionName,
				URL:  fmt.Sprintf("https://pokeapi.co/api/v2/region/%s", regionName),
			},
		}

		// Serialize the expected location to JSON (this is what the cache would store)
		rawData, err := json.Marshal(expectedLocation)
		if err != nil {
			t.Fatalf("failed to marshal location: %v", err)
		}

		// Create a cache and pre-populate it with the expected URL key
		cache := pokecache.NewCache(60000) // 60 second TTL
		expectedURL := fmt.Sprintf("https://pokeapi.co/api/v2/location/%s", locationName)
		cache.Add(expectedURL, rawData)

		// Call GetLocation — it should return cached data without making an HTTP request
		result, err := GetLocation(locationName, cache)
		if err != nil {
			t.Fatalf("GetLocation returned error for cached entry: %v", err)
		}

		// Verify the returned Location matches what was cached
		if result.ID != expectedLocation.ID {
			t.Errorf("ID mismatch: got %d, want %d", result.ID, expectedLocation.ID)
		}
		if result.Name != expectedLocation.Name {
			t.Errorf("Name mismatch: got %q, want %q", result.Name, expectedLocation.Name)
		}
		if len(result.Areas) != len(expectedLocation.Areas) {
			t.Errorf("Areas count mismatch: got %d, want %d", len(result.Areas), len(expectedLocation.Areas))
		}
		if result.Region.Name != expectedLocation.Region.Name {
			t.Errorf("Region mismatch: got %q, want %q", result.Region.Name, expectedLocation.Region.Name)
		}

		// Verify the cache still contains the URL key (Requirement 5.1)
		_, ok := cache.Get(expectedURL)
		if !ok {
			t.Errorf("cache entry missing for URL key %q after GetLocation call", expectedURL)
		}
	})
}
