package utils

import (
	"os"
	"strings"
)

type Settings struct {
	Api_token string
	Debug     bool
	Raw       map[string]string
	TabHeight int
}

func GetEnvironMap() map[string]string {
	var rawEnviron = os.Environ()
	var mapEnviron = map[string]string{}

	for rawEnvLine := range rawEnviron {
		envPair := strings.Split(rawEnviron[rawEnvLine], "=")
		if len(envPair) == 2 {
			mapEnviron[envPair[0]] = envPair[1]
		}
	}

	return mapEnviron
}

// var environ = GetEnvironMap()

func GetSettings() *Settings {
	var raw = GetEnvironMap()

	api_token := raw["token"]
	_, debug := raw["debug"]

	var settings = Settings{
		Api_token: api_token,
		Debug:     debug,
		Raw:       raw,
		TabHeight: 3,
	}

	return &settings
}
