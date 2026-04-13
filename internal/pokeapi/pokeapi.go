package pokeapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"pokedex-go/internal/pokecache"
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

type LocationAreaDetails struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	GameIndex            int    `json:"game_index"`
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	Location struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Names []struct {
		Name     string `json:"name"`
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
			MaxChance        int `json:"max_chance"`
			EncounterDetails []struct {
				MinLevel        int   `json:"min_level"`
				MaxLevel        int   `json:"max_level"`
				ConditionValues []any `json:"condition_values"`
				Chance          int   `json:"chance"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
			} `json:"encounter_details"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

func GetLocationAreas(url string, c *pokecache.Cache) (LocationAreas, error) {
	
	var locationAreas LocationAreas

	val, ok := c.Get(url)
	if ok {
		buf := bytes.NewBuffer(val)
		json.NewDecoder(buf).Decode(&locationAreas)
		return locationAreas, nil
	}
	
	res, err := http.Get(url)
	if err != nil {
		return LocationAreas{}, fmt.Errorf("Error getting Location Area: %v\n", err)
	}
	defer res.Body.Close()
	
	json.NewDecoder(res.Body).Decode(&locationAreas)
	rawData, _ := json.Marshal(locationAreas)
	c.Add(url, rawData)

	return locationAreas, nil

}

func GetLocationAreaDetails(locationName string, c *pokecache.Cache) (LocationAreaDetails, error) {
	
	var locationAreaDetails LocationAreaDetails
	val, ok := c.Get(fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", locationName))
	if ok {
		buf := bytes.NewBuffer(val)
		json.NewDecoder(buf).Decode(&locationAreaDetails)
		return locationAreaDetails, nil
	}
	

	res, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", locationName))
	if err != nil {
		return LocationAreaDetails{}, fmt.Errorf("Error getting Location Area Details: %v\n", err)
	}
	defer res.Body.Close()
	
	json.NewDecoder(res.Body).Decode(&locationAreaDetails)
	rawData, _ := json.Marshal(locationAreaDetails)
	c.Add(fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", locationName), rawData)

	return locationAreaDetails, nil
}

