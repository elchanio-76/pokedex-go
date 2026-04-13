package main

import (
	_ "bufio"
	"fmt"
	"os"

	"pokedex-go/internal/input"
	"pokedex-go/internal/pokeapi"
	"pokedex-go/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(cfg *Config, args []string, t input.Terminal) error
}

type Config struct {
	Next string
	Prev string
	Pokedex map[string]string
	commandCache []string
	historyIndex int
}

func commandExit(cfg *Config, args []string, t input.Terminal) error {
	t.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config, args []string, t input.Terminal) error {
	t.Print("Welcome to the Pokedex!\nUsage:\n\n")
	for _, cmd := range commandRegistry {
		t.Print("  %s: %s\n", cmd.name, cmd.description)
	}
	return nil
}


func commandMap(cfg *Config, args []string, t input.Terminal) error {
	if cfg.Next == "" {
		t.Println("You're on the last page!")
		return nil
	}

	res, err := pokeapi.GetLocationAreas(cfg.Next, cache)
	if err != nil {
		t.Print("Error fetching data: %s\n", err)
		return err
	}
	
	for _, area := range res.Results {
		t.Print("  %s\n", area.Name)
	}
	cfg.Next = res.Next
	cfg.Prev = res.Previous

	return nil
}

func commandMapBack(cfg *Config, args []string, t input.Terminal) error {
	if cfg.Prev == "" {
		t.Println("You're on the first page!")
		return nil
	}
	res, err := pokeapi.GetLocationAreas(cfg.Prev, cache)
	if err != nil {
		t.Print("Error fetching data: %s\n", err)
		return err
	}

	for _, area := range res.Results {
		t.Print("  %s\n", area.Name)
	}
	cfg.Next = res.Next
	cfg.Prev = res.Previous
	return nil
}

func commandExplore(cfg *Config, args []string, t input.Terminal) error {
	if len(args) != 1 {
		t.Println("You must provide a location area to explore")
		return nil
	}
	
	res, err := pokeapi.GetLocationAreaDetails(args[0], cache)
	if err != nil {
		t.Print("Error fetching data: %s\n", err)
		return err
	}

	t.Print("Pokemons in %s:\n", args[0])
	for _, pokemon := range res.PokemonEncounters {
		t.Print("  - %s\n", pokemon.Pokemon.Name)
	}
	return nil
}

func commandCatch(cfg *Config, args []string, t input.Terminal) error {
	if len(args) != 1 {
		t.Println("You must provide a pokemon to catch")
		return nil
	}
	_, exists := cfg.Pokedex[args[0]]
	if exists {
		return fmt.Errorf("%s already in Pokedex!", args[0])
	}

	caught, err := pokeapi.CatchPokemon(args[0], cache)
	if err != nil {
		t.Print("Error fetching data: %s\n", err)
		return err
	}

	t.Print("Throwing a Pokeball at %s...\n", args[0])
	if caught {
		cfg.Pokedex[args[0]] = args[0]
		t.Print("%s was caught!\n", args[0])
	} else {
		t.Print("%s was not caught! Try again.\n", args[0])
	}
	
	return nil
}

func commandInspect(cfg *Config, args []string, t input.Terminal) error {
	if len(args) != 1 {
		t.Println("You must provide a pokemon to inspect")
		return nil
	}
	pokemon, ok := cfg.Pokedex[args[0]]
	if !ok {
		return fmt.Errorf("%s is not in your pokedex", args[0])
	}

	details, err := pokeapi.GetPokemonDetails(pokemon, cache)
	if err != nil {
		t.Print("Error fetching data: %s\n", err)
		return err
	}

	t.Print("Name: %s\n", details.Name)
	t.Print("Height: %d\n", details.Height)
	t.Print("Weight: %d\n", details.Weight)
	t.Print("Stats:\n")
	for _, stat := range details.Stats {
		t.Print("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	t.Print("Types:\n")
	for _, typ := range details.Types {
		t.Print("  -%s\n", typ.Type.Name)
	}
	return nil
}

func commandPokedex(cfg *Config, args []string, t input.Terminal) error {
	t.Println("Your Pokedex:")
	for p := range cfg.Pokedex {
		t.Print(" - %s\n", p)
	}
	t.Println("End of Pokedex")
	return nil
}

type CommandRegistry map[string]cliCommand

var commandRegistry CommandRegistry
var cache = pokecache.NewCache(5 * 60 * 1000)

func init() {
	commandRegistry = CommandRegistry{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "List the next 20 locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "List the previous 20 locations",
			callback:    commandMapBack,
		},
		"explore": {
			name:        "explore",
			description: "Explore a location",
			callback:    commandExplore,
		},
		"catch": {
			name:					"catch",
			description:	"Catch a pokemon",
			callback:			commandCatch,
		},
		"inspect": {
			name:					"inspect",
			description:	"Inspect caught pokemon stats",
			callback:			commandInspect,
		},
		"pokedex": {
			name:					"pokedex",
			description:	"Show your Pokedex",
			callback:			commandPokedex,
		},
	}
}

func main() {
	
	cfg := Config{
		Next: "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
		Prev: "",
		Pokedex: make(map[string]string),
		commandCache: []string{},
		historyIndex: 0,
	}
	var reader input.Terminal = input.NewTerminalReader()
	startREPL(&cfg, reader)
	
}


