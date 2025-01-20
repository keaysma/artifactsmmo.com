package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
)

type GetAllMonstersParams struct {
	Drop *string
}

func GetMonsterByCode(code string) (*types.Monster, error) {
	var out types.Monster
	err := GetDataResponseFuture(
		fmt.Sprintf("monsters/%s", code),
		nil,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

func GetAllMonsters(in GetAllMonstersParams) (*[]types.Monster, error) {
	var payload = map[string]string{}
	if in.Drop != nil {
		payload["drop"] = *in.Drop
	}

	var out []types.Monster
	err := GetDataResponseFuture(
		"monsters",
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
