package main

import (
	"fmt"
	"strings"
	_"bufio"
	_"os"

	"pokedex-go/internal/input"
)

func startREPL(cfg *Config, reader input.LineReader) {
	defer reader.Close()
	reader.SetHistory(cfg.commandCache)

	for {
		text, err := reader.ReadLine("Pokedex > ")
		if err != nil {
			break
		}
		cfg.commandCache = append(cfg.commandCache, text)
		words := cleanInput(text)
		cmd, ok := commandRegistry[words[0]]
		if !ok {
			fmt.Printf("Unknown command %s\n", words[0])
			continue
		}
		args := words[1:]
		err = cmd.callback(cfg, args)
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