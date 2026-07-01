package main

import (
	_ "bufio"
	"fmt"
	"os"

	"pokedex-go/internal/input"
	"pokedex-go/internal/pokeapi"
	"pokedex-go/internal/pokecache"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Magenta = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

const MaxNameLen = 30

func truncateName(name string, maxLen int) string {
	if len(name) > maxLen {
		return name[:maxLen-3] + "..."
	}
	return name
}

type cliCommand struct {
	name        string
	description string
	callback    func(cfg *Config, args []string, t input.Terminal) error
}

type Config struct {
	Next            string
	Prev            string
	Pokedex         map[string]string
	CurrentLocation *pokeapi.Location
	CurrentArea     string
	ScopedPageIndex int
	commandCache    []string
	historyIndex    int
}

func commandExit(cfg *Config, args []string, t input.Terminal) error {
	t.Println("Closing the Pokedex... Goodbye!")
	t.Close()
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
		t.Println(Red + "You're on the first page!" + Reset)
		return nil
	}
	res, err := pokeapi.GetLocationAreas(cfg.Prev, cache)
	if err != nil {
		t.Print(Red+"Error fetching data: %s\n"+Reset, err)
		return err
	}

	for _, area := range res.Results {
		t.Print(Blue+"  %s\n"+Reset, area.Name)
	}
	cfg.Next = res.Next
	cfg.Prev = res.Previous
	return nil
}

func commandExplore(cfg *Config, args []string, t input.Terminal) error {
	if len(args) != 1 {
		t.Println(Red + "You must provide a location area to explore" + Reset)
		return nil
	}

	res, err := pokeapi.GetLocationAreaDetails(args[0], cache)
	if err != nil {
		t.Print(Red+"Error fetching data: %s\n"+Reset, err)
		return err
	}

	t.Print("Pokemons in %s:\n", args[0])
	for _, pokemon := range res.PokemonEncounters {
		t.Print(Blue+"  - %s\n"+Reset, pokemon.Pokemon.Name)
	}
	return nil
}

func commandCatch(cfg *Config, args []string, t input.Terminal) error {
	if len(args) != 1 {
		t.Println(Red + "You must provide a pokemon to catch" + Reset)
		return nil
	}
	_, exists := cfg.Pokedex[args[0]]
	if exists {
		return fmt.Errorf(Red+"%s already in Pokedex!"+Reset, args[0])
	}

	caught, err := pokeapi.CatchPokemon(args[0], cache)
	if err != nil {
		t.Print(Red+"Error fetching data: %s\n"+Reset, err)
		return err
	}

	t.Print(Yellow+"Throwing a Pokeball at %s...\n"+Reset, args[0])
	if caught {
		cfg.Pokedex[args[0]] = args[0]
		t.Print(Green+"%s was caught!\n"+Reset, args[0])
	} else {
		t.Print(Magenta+"%s was not caught! Try again.\n"+Reset, args[0])
	}

	return nil
}

func commandInspect(cfg *Config, args []string, t input.Terminal) error {
	if len(args) != 1 {
		t.Println(Red + "You must provide a pokemon to inspect" + Reset)
		return nil
	}
	pokemon, ok := cfg.Pokedex[args[0]]
	if !ok {
		return fmt.Errorf(Red+"%s is not in your pokedex"+Reset, args[0])
	}

	details, err := pokeapi.GetPokemonDetails(pokemon, cache)
	if err != nil {
		t.Print(Red+"Error fetching data: %s\n"+Reset, err)
		return err
	}

	t.Print("Name: %s\n", details.Name)
	t.Print("Height: %d\n", details.Height)
	t.Print("Weight: %d\n", details.Weight)
	t.Print("Stats:\n")
	for _, stat := range details.Stats {
		t.Print(Green+"  -%s: %d\n"+Reset, stat.Stat.Name, stat.BaseStat)
	}
	t.Print("Types:\n")
	for _, typ := range details.Types {
		t.Print(Cyan+"  -%s\n"+Reset, typ.Type.Name)
	}
	return nil
}

func commandPokedex(cfg *Config, args []string, t input.Terminal) error {
	t.Println("Your Pokedex:")
	for p := range cfg.Pokedex {
		t.Print(Blue+" - %s\n"+Reset, p)
	}
	t.Println("End of Pokedex")
	return nil
}

func commandVisit(cfg *Config, args []string, t input.Terminal) error {
	if len(args) != 1 {
		t.Println("You must provide a location name to visit")
		return nil
	}

	loc, err := pokeapi.GetLocation(args[0], cache)
	if err != nil {
		t.Print("Location '%s' not found\n", args[0])
		return nil
	}

	cfg.CurrentLocation = &loc
	cfg.CurrentArea = ""
	cfg.ScopedPageIndex = 0

	t.Print(Green+"Location: %s\n"+Reset, loc.Name)
	t.Print(Cyan+"Region: %s\n"+Reset, loc.Region.Name)
	t.Print(Yellow+"Areas: %d\n"+Reset, len(loc.Areas))

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
			name:        "catch",
			description: "Catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect caught pokemon stats",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Show your Pokedex",
			callback:    commandPokedex,
		},
		"visit": {
			name:        "visit",
			description: "Visit a location to explore its areas",
			callback:    commandVisit,
		},
	}
}

func main() {

	cfg := Config{
		Next:         "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
		Prev:         "",
		Pokedex:      make(map[string]string),
		commandCache: []string{},
		historyIndex: 0,
	}
	var reader input.Terminal = input.NewTerminalReader()
	startREPL(&cfg, reader)

}
