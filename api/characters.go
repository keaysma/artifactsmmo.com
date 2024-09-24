package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

// https://api.artifactsmmo.com/docs/#/operations/get_character_characters__name__get

func GetCharacterByName(name string) (*types.Character, error) {
	res, err := GetDataResponse(fmt.Sprintf("characters/%s", name), nil)

	if err != nil {
		return nil, err
	}

	var out types.Character
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, err
	}

	return &out, nil
}

func GetAllCharacters() (*[]types.Character, error) {
	res, err := GetDataResponse("characters", nil)
	if err != nil {
		return nil, err
	}

	var out []types.Character
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, err
	}

	return &out, nil
}
