package utils

import (
	"os"
	"strings"
)

type Settings struct {
	Api_token      string
	Start_commands []string
	Debug          bool
	Raw            map[string]string
	TabHeight      int
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

	start_commands := []string{}
	raw_start_commands, has := raw["start_commands"]
	if has {
		start_commands = strings.Split(raw_start_commands, ",")
	}

	var settings = Settings{
		Api_token:      api_token,
		Start_commands: start_commands,
		Debug:          debug,
		Raw:            raw,
		TabHeight:      3,
	}

	return &settings
}
