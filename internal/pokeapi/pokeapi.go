package pokeapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type LocationAreas struct{

	Count 		string `json:"count"`	
	Results []struct{
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
	Next 	string `json:"next"`
	Previous string `json:"previous"`
}

func GetLocationAreas(url string) (LocationAreas, error) {
	res, err := http.Get(url)
	if err != nil {
		return LocationAreas{}, fmt.Errorf("Error getting Location Area: %v\n", err)
	}
	defer res.Body.Close()
	var locationAreas LocationAreas
	json.NewDecoder(res.Body).Decode(&locationAreas)

	return locationAreas, nil

}