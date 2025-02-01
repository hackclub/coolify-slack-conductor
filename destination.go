package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"regexp"
)

type ConfigItem struct {
	Name  string
	Regex []string
}
type Config struct {
	Destinations []ConfigItem
}

type Destination struct {
	url    string
	regexp []string
}

func (dest Destination) Matches(body string) bool {
	for _, regex := range dest.regexp {
		match, _ := regexp.MatchString(regex, body)
		if match {
			return true
		}
	}
	return false
}

var MainDestination = Destination{
	url: os.Getenv("WEBHOOK_MAIN_URL"),
}

var LoadedConfig Config
var LoadedDestinations []Destination
var ConfigIsLoaded = false

func loadDestinations() {
	if ConfigIsLoaded {
		return
	}

	config, _ := os.ReadFile("config.yml")
	err := yaml.Unmarshal(config, &LoadedConfig)
	if err != nil {
		panic(err)
	}
	fmt.Printf("loaded config:\n%v\n\n", LoadedConfig)

	// Convert config into destinations
	for _, item := range LoadedConfig.Destinations {
		envVar := fmt.Sprintf("WEBHOOK_%s_URL", item.Name)
		dest := Destination{
			url:    os.Getenv(envVar),
			regexp: item.Regex,
		}
		if dest.url == "" {
			log.Fatal(fmt.Sprintf("Missing %s environment variable", envVar))
		}

		LoadedDestinations = append(LoadedDestinations, dest)
	}

	ConfigIsLoaded = true
}

func destinations(body string) []Destination {
	var results []Destination
	for _, dest := range LoadedDestinations {
		if dest.Matches(body) {
			results = append(results, dest)
		}
	}
	log.Println("Destinations:", results)
	return results
}
