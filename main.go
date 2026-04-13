package main

import (
	"bufio"
	"fmt"
	"os"

	"pokedex-go/internal/pokeapi"
	"pokedex-go/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(cfg *Config, args []string) error
}

type Config struct {
	Next string
	Prev string
}

func commandExit(cfg *Config, args[] string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config, args [] string) error {
	fmt.Printf("Welcome to the Pokedex!\nUsage:\n\n")
	for _, cmd := range commandRegistry {
		fmt.Printf("  %s: %s\n", cmd.name, cmd.description)
	}
	return nil
}


func commandMap(cfg *Config, args []string) error {
	if cfg.Next == "" {
		fmt.Println("You're on the last page!")
		return nil
	}

	res, err := pokeapi.GetLocationAreas(cfg.Next, cache)
	if err != nil {
		fmt.Printf("Error fetching data: %s\n", err)
		return err
	}
	
	for _, area := range res.Results {
		fmt.Printf("  %s\n", area.Name)
	}
	cfg.Next = res.Next
	cfg.Prev = res.Previous

	return nil
}

func commandMapBack(cfg *Config, args []string) error {
	if cfg.Prev == "" {
		fmt.Println("You're on the first page!")
		return nil
	}
	res, err := pokeapi.GetLocationAreas(cfg.Prev, cache)
	if err != nil {
		fmt.Printf("Error fetching data: %s\n", err)
		return err
	}

	for _, area := range res.Results {
		fmt.Printf("  %s\n", area.Name)
	}
	cfg.Next = res.Next
	cfg.Prev = res.Previous
	return nil
}

func commandExplore(cfg *Config, args []string) error {
	if len(args) != 1 {
		fmt.Println("You must provide a location area to explore")
		return nil
	}
	
	res, err := pokeapi.GetLocationAreaDetails(args[0], cache)
	if err != nil {
		fmt.Printf("Error fetching data: %s\n", err)
		return err
	}

	fmt.Printf("Pokemons in %s:\n", args[0])
	for _, pokemon := range res.PokemonEncounters {
		fmt.Printf("  - %s\n", pokemon.Pokemon.Name)
	}
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
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	cfg := Config{
		Next: "https://pokeapi.co/api/v2/location-area?offset=0&limit=20",
		Prev: "",
	}

	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		words := cleanInput(text)
		cmd, ok := commandRegistry[words[0]]
		if !ok {
			fmt.Printf("Unknown command %s\n", words[0])
			continue
		}
		args := words[1:]
		err := cmd.callback(&cfg, args)
		if err != nil {
			fmt.Println(err)
		}
	}
}


