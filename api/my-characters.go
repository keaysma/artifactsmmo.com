package api

import (
	"artifactsmmo.com/m/types"
)

func GetAllMyCharacters() (*[]types.Character, error) {
	var out []types.Character
	err := GetDataResponse(
		"my/characters",
		nil,
		&out,
	)

	if err != nil {
		return nil, err
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

type GetLogsParams struct {
	Page int
	Size int
}

func GetLogs(in GetLogsParams) (*[]Log, error) {
	var out []Log
	err := GetDataResponse(
		"my/logs",
		in,
		&out,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}
