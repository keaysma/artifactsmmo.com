package api

import (
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

func GetAllMyCharacters() (*[]types.Character, error) {
	res, err := GetDataResponse("my/characters", nil)
	if err != nil {
		return nil, err
	}

	var out []types.Character
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}

// https://api.artifactsmmo.com/docs/#/operations/get_all_characters_logs_my_logs_get

type Log struct {
	Character           string
	Account             string
	Type                string
	Description         string
	Content             interface{} // what is this?
	Cooldown            int
	Cooldown_expiration string
	Created_at          string
}

func GetLogs(page int, limit int) (*[]Log, error) {
	res, err := GetDataResponse("my/logs", &map[string]string{
		"page":  string(rune(page)),
		"limit": string(rune(limit)),
	})
	if err != nil {
		return nil, err
	}

	var out []Log
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
