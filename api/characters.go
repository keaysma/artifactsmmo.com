package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
)

// https://api.artifactsmmo.com/docs/#/operations/get_character_characters__name__get

func GetCharacterByName(name string) (*types.Character, error) {
	var out types.Character
	err := GetDataResponseFuture(
		fmt.Sprintf("characters/%s", name),
		nil,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

func GetAllCharacters() (*[]types.Character, error) {
	var out []types.Character
	err := GetDataResponseFuture(
		"my/characters",
		nil,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
