# Pokedex Go

A terminal-based Pokédex application written in Go. Explore locations, discover Pokémon, catch them, and build your collection — all from the command line.

## Features

- Browse and visit locations from the Pokémon world
- Explore location areas to find wild Pokémon encounters
- Catch Pokémon with probability based on base experience
- Inspect caught Pokémon stats (height, weight, types, abilities)
- Maintain a personal Pokédex of caught Pokémon
- In-memory cache with automatic expiration to reduce API calls
- Interactive REPL with command history and colored output

## Requirements

- Go 1.25+

## Getting Started

```bash
# Clone the repository
git clone <repo-url>
cd pokedex-go

# Run the application
go run .
```

## Commands

| Command | Description |
|---------|-------------|
| `help` | Display available commands |
| `showloc` | List the next 20 locations |
| `showlocb` | List the previous 20 locations |
| `visit <name>` | Visit a location to explore its areas |
| `map` | List areas in the current location (paginated) |
| `mapb` | Go back a page of areas |
| `explore <area>` | Show Pokémon encounters in an area |
| `catch <pokemon>` | Attempt to catch a Pokémon |
| `inspect <pokemon>` | View stats of a caught Pokémon |
| `pokedex` | Show all caught Pokémon |
| `exit` | Exit the application |

## Example Session

```
Pokedex > showloc
  canalave-city
  eterna-city
  pastoria-city
  ...

Pokedex > visit eterna-city
Location: eterna-city
Region: sinnoh
Areas: 2

[Location: eterna-city] Pokedex > map
  eterna-city-area
  eterna-city-west-gate-area

[Location: eterna-city] Pokedex > explore eterna-city-area
Pokemons in eterna-city-area:
  - buneary
  - wurmple
  - ...

[Location: eterna-city | Area: eterna-city-area] Pokedex > catch buneary
Throwing a Pokeball at buneary...
buneary was caught!

[Location: eterna-city | Area: eterna-city-area] Pokedex > inspect buneary
Name: buneary
Height: 4
Weight: 55
Stats:
  -hp: 55
  -attack: 66
  ...
Types:
  -normal
```

## Project Structure

```
.
├── main.go                  # Entry point, command definitions
├── repl.go                  # REPL loop and prompt logic
├── internal/
│   ├── input/               # Terminal input abstraction
│   ├── pokeapi/             # PokeAPI client with caching
│   └── pokecache/           # Thread-safe in-memory cache
├── *_test.go                # Tests
├── go.mod
└── go.sum
```

## API

This project uses the [PokéAPI](https://pokeapi.co/) — a free, open RESTful API for Pokémon data.

## Running Tests

```bash
go test ./...
```
