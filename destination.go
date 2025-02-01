package main

import "regexp"

type Destination struct {
	url    string
	regexp string
}

var MainDestination = Destination{
	url: "https://webhook.site/a32c93b6-747e-4f29-8b2c-df6eb59b6fb2",
}

func (dest Destination) Matches(body string) bool {
	match, _ := regexp.MatchString(dest.regexp, body)
	return match
}

func destinations(body string) []Destination {
	config := []Destination{
		{
			url:    "https://webhook.site/a32c93b6-747e-4f29-8b2c-df6eb59b6fb2?afm",
			regexp: "\\*\\*Project:\\*\\* gary@mfa\\\\n",
		},
		{
			url:    "https://webhook.site/a32c93b6-747e-4f29-8b2c-df6eb59b6fb2?danlog",
			regexp: "goland",
		},
	}

	var results []Destination
	for _, dest := range config {
		if dest.Matches(body) {
			results = append(results, dest)
		}
	}
	return results
}
