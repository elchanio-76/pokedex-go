package main

import (
	_ "bufio"
	_ "os"
	"strings"

	"pokedex-go/internal/input"
)

func startREPL(cfg *Config, reader input.Terminal) {
	defer reader.Close()
	reader.SetHistory(cfg.commandCache)

	for {
		text, err := reader.ReadLine("Pokedex > ")
		if err != nil {
			break
		}
		cfg.commandCache = append(cfg.commandCache, text)
		cfg.historyIndex = len(cfg.commandCache) // set history index to last command

		words := cleanInput(text)
		cmd, ok := commandRegistry[words[0]]
		if !ok {
			reader.Print("Unknown command %s\n", words[0])
			continue
		}
		args := words[1:]
		err = cmd.callback(cfg, args, reader)
		if err != nil {
			reader.Println(err)
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