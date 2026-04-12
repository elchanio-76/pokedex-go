package main

import (
	"bufio"
	"fmt"
	"os"

	"pokedex-go/internal/pokeapi"
)

type cliCommand struct {
	name        string
	description string
	callback    func(cfg *Config) error
}

type Config struct {
	Next string
	Prev string
}

func commandExit(cfg *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config) error {
	fmt.Printf("Welcome to the Pokedex!\nUsage:\n\n")
	for _, cmd := range commandRegistry {
		fmt.Printf("  %s: %s\n", cmd.name, cmd.description)
	}
	return nil
}


func commandMap(cfg *Config) error {
	if cfg.Next == "" {
		fmt.Println("You're on the last page!")
		return nil
	}
	
	res, err := pokeapi.GetLocationAreas(cfg.Next)
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

func commandMapBack(cfg *Config) error {
	if cfg.Prev == "" {
		fmt.Println("You're on the first page!")
		return nil
	}
	res, err := pokeapi.GetLocationAreas(cfg.Prev)
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

type CommandRegistry map[string]cliCommand

var commandRegistry CommandRegistry

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
		err := cmd.callback(&cfg)
		if err != nil {
			fmt.Println(err)
		}
	}
}


