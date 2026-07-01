package main

import (
	"strings"
	"testing"

	"pokedex-go/internal/pokeapi"

	"pgregory.net/rapid"
)

// TestProperty_PromptBuilding validates Property 8: Prompt Building.
// **Validates: Requirements 4.1, 4.2, 4.3**
func TestProperty_PromptBuilding(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate random state: 0 = nil location, 1 = location only, 2 = location + area
		state := rapid.IntRange(0, 2).Draw(t, "state")
		cfg := &Config{Pokedex: make(map[string]string)}

		switch state {
		case 0:
			cfg.CurrentLocation = nil
			cfg.CurrentArea = ""
		case 1:
			locName := rapid.StringN(1, 50, 50).Draw(t, "locName")
			cfg.CurrentLocation = &pokeapi.Location{Name: locName}
			cfg.CurrentArea = ""
		case 2:
			locName := rapid.StringN(1, 50, 50).Draw(t, "locName")
			areaName := rapid.StringN(1, 50, 50).Draw(t, "areaName")
			cfg.CurrentLocation = &pokeapi.Location{Name: locName}
			cfg.CurrentArea = areaName
		}

		result := buildPrompt(cfg)

		// Property: always ends with "Pokedex > "
		if !strings.HasSuffix(result, "Pokedex > ") {
			t.Fatalf("prompt does not end with 'Pokedex > ': %q", result)
		}

		switch state {
		case 0:
			if result != "Pokedex > " {
				t.Fatalf("expected 'Pokedex > ' for nil location, got %q", result)
			}
		case 1:
			if !strings.Contains(result, "[Location:") {
				t.Fatalf("expected '[Location:' in prompt, got %q", result)
			}
			if strings.Contains(result, "| Area:") {
				t.Fatalf("did not expect '| Area:' in location-only prompt, got %q", result)
			}
		case 2:
			if !strings.Contains(result, "[Location:") {
				t.Fatalf("expected '[Location:' in prompt, got %q", result)
			}
			if !strings.Contains(result, "| Area:") {
				t.Fatalf("expected '| Area:' in prompt, got %q", result)
			}
		}
	})
}
