package main

import (
	"fmt"
	"strings"
	"bufio"
	"os"
)

func startREPL(cfg *Config) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()
		cfg.commandCache = append(cfg.commandCache, text)
		words := cleanInput(text)
		cmd, ok := commandRegistry[words[0]]
		if !ok {
			fmt.Printf("Unknown command %s\n", words[0])
			continue
		}
		args := words[1:]
		err := cmd.callback(cfg, args)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func cleanInput(text string) []string {
	var result []string

	words := strings.Split(text, " ")
	for i:=range words {
		if strings.Trim(words[i]," ")=="" {
			continue
		}

		result = append(result, strings.ToLower(strings.Trim(words[i]," ")))
	}
	
	return result
}