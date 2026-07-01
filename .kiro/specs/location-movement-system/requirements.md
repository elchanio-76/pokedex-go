# Requirements Document

## Introduction

This feature introduces a Location layer above LocationAreas in the Pokédex REPL application, creating a navigation/movement system. Users will "visit" a Location (e.g., a city or route) and then browse or explore only the LocationAreas belonging to that Location. The REPL prompt will dynamically reflect the user's current location context using ANSI colors.

## Glossary

- **REPL**: The Read-Eval-Print Loop that drives the CLI application, accepting user commands and displaying output
- **Location**: A named place in the Pokémon world (city, route, etc.) fetched from `https://pokeapi.co/api/v2/location/<name>`. Contains a list of LocationAreas
- **LocationArea**: A specific sub-area within a Location where Pokémon encounters occur, fetched from `https://pokeapi.co/api/v2/location-area/<name>`
- **Config**: The application state struct that holds navigation state, Pokédex data, and command history
- **Terminal**: The interface providing ReadLine, Print, Println, SetHistory, and Close methods for terminal I/O
- **Prompt**: The text displayed before user input in the REPL, currently hardcoded as "Pokedex > "
- **PokeAPI**: The external REST API at `pokeapi.co` providing Pokémon game data
- **Cache**: The TTL-based in-memory cache used to avoid redundant HTTP requests

## Requirements

### Requirement 1: Visit a Location

**User Story:** As a user, I want to visit a Location by name, so that I can set my current position and browse its areas.

#### Acceptance Criteria

1. WHEN the user executes `visit <location_name>`, THE REPL SHALL send a GET request to `https://pokeapi.co/api/v2/location/<location_name>` and store the response as the current Location in Config
2. WHEN the visit command succeeds, THE REPL SHALL display the Location name, the region name, and the total number of areas (derived from the length of the `areas` array in the response)
3. IF the PokeAPI returns an error or a non-200 HTTP status for the visit request, THEN THE REPL SHALL display an error message including the location name that was not found, and SHALL NOT modify the current Location in Config
4. WHEN the user executes `visit` without providing a location name argument, THE REPL SHALL display an error message indicating that a location name is required
5. WHEN the user visits a new Location successfully, THE REPL SHALL replace the previously stored Location in Config with the new one and set the currently selected LocationArea to nil
6. WHEN the visit command sends a GET request, THE REPL SHALL use the existing TTL cache to store and retrieve the HTTP response, returning the cached response for repeated visits to the same location within the cache TTL

### Requirement 2: Scoped Map Navigation

**User Story:** As a user, I want the `map` and `mapb` commands to list LocationAreas within my current Location, so that I can browse areas relevant to where I am.

#### Acceptance Criteria

1. WHILE a Location is set in Config, WHEN the user executes `map`, THE REPL SHALL display the next page of LocationArea names from the current Location's areas list, starting from page index 0 on the first invocation after a Location is set or changed
2. WHILE a Location is set in Config, WHEN the user executes `mapb`, THE REPL SHALL display the previous page of LocationArea names from the current Location's areas list
3. WHEN the user executes `map` and the current page is already the last page of the areas list, THE REPL SHALL display a message indicating the user is on the last page and not advance the page index
4. WHEN the user executes `mapb` and the current page index is 0, THE REPL SHALL display a message indicating the user is on the first page and not decrement the page index
5. IF no Location is currently set in Config, THEN THE REPL SHALL display an error message instructing the user to visit a location first when `map` or `mapb` is executed
6. THE REPL SHALL paginate the Location's areas list in groups of 20 items per page
7. IF a Location is set in Config and its areas list is empty, THEN THE REPL SHALL display a message indicating the current Location has no areas to show when `map` or `mapb` is executed
8. WHEN a new Location is set in Config, THE REPL SHALL reset the scoped map page index to 0

### Requirement 3: Scoped Explore Command

**User Story:** As a user, I want the `explore` command to only accept LocationAreas belonging to my current Location, so that navigation stays consistent.

#### Acceptance Criteria

1. WHILE a Location is set in Config, WHEN the user executes `explore <area_name>` with an area whose name matches an entry in the current Location's `areas` list, THE REPL SHALL call the location-area API for that area and display each Pokémon encounter name on a separate line prefixed with "- "
2. WHILE a Location is set in Config, WHEN the user executes `explore <area_name>` with an area name that does not match any entry in the current Location's `areas` list, THE REPL SHALL display an error message indicating the area is not part of the current Location and SHALL NOT make an API call for that area
3. IF no Location is currently set in Config, THEN THE REPL SHALL display an error message instructing the user to visit a location first when `explore <area_name>` is executed
4. WHEN the REPL successfully fetches and displays encounters for an area, THE REPL SHALL store that area name as the currently active LocationArea in Config
5. IF the location-area API call fails after validation passes, THEN THE REPL SHALL display an error message indicating the fetch failure and SHALL NOT update the currently active LocationArea in Config

### Requirement 4: Dynamic Colored Prompt

**User Story:** As a user, I want the REPL prompt to show my current Location and active area in color, so that I always know my navigation context.

#### Acceptance Criteria

1. WHILE a Location is set in Config and no LocationArea is active, THE REPL SHALL display the prompt as `[Location: <location_name>] Pokedex > ` with the bracketed location segment rendered using the Cyan ANSI color code followed by Reset before the `Pokedex > ` portion
2. WHILE both a Location and a LocationArea are active in Config, THE REPL SHALL display the prompt as `[Location: <location_name> | Area: <area_name>] Pokedex > ` with the bracketed context segment rendered using the Cyan ANSI color code followed by Reset before the `Pokedex > ` portion
3. WHILE no Location is set in Config, THE REPL SHALL display the default prompt `Pokedex > ` with no ANSI color codes applied
4. WHEN the user visits a new Location, THE REPL SHALL update the prompt to reflect the new Location name starting from the next input line
5. WHEN the user explores an area, THE REPL SHALL update the prompt to include the area name starting from the next input line
6. IF the location_name or area_name exceeds 30 characters, THEN THE REPL SHALL truncate the displayed name to 27 characters followed by `...` in the prompt segment

### Requirement 5: Location Data Caching

**User Story:** As a user, I want Location API responses to be cached, so that revisiting the same location does not require another network request.

#### Acceptance Criteria

1. WHEN the REPL fetches a Location from PokeAPI and receives a successful response, THE Cache SHALL store the JSON-encoded response body keyed by the full request URL (e.g., `https://pokeapi.co/api/v2/location/{name}`)
2. WHEN the user requests a Location that exists in the Cache and the entry has not exceeded the TTL, THE REPL SHALL return the cached data without making an HTTP request to PokeAPI
3. THE Cache SHALL apply the same TTL expiration policy (5 minutes) to Location entries as it does to existing LocationArea entries
4. IF a cached Location entry has exceeded the 5-minute TTL, THEN THE Cache SHALL discard the expired entry and THE REPL SHALL fetch the Location from PokeAPI via a new HTTP request

### Requirement 6: Register Visit Command

**User Story:** As a developer, I want the `visit` command registered in the CommandRegistry, so that the REPL recognizes it alongside existing commands.

#### Acceptance Criteria

1. THE CommandRegistry SHALL contain a `visit` entry with name "visit", a description string of "Visit a location to explore its areas", and a callback function with signature `func(cfg *Config, args []string, t input.Terminal) error`
2. WHEN the user executes `help`, THE REPL SHALL include the `visit` command and its description in the printed list of available commands
