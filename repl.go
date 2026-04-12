package main

import (
	_ "fmt"
	"strings"
)



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