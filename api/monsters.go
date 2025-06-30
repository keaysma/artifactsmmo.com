package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
)

type GetAllMonstersParams struct {
	Drop *string
}

// sos being too hacky
var CACHE_THAT_SHOULD_BE_DONE_BETTER_AND_ELSEWHERE = map[string]*types.Monster{}

func GetMonsterByCode(code string) (*types.Monster, error) {
	cached, inCache := CACHE_THAT_SHOULD_BE_DONE_BETTER_AND_ELSEWHERE[code]
	if inCache {
		return cached, nil
	}

	var out types.Monster
	err := GetDataResponseFuture(
		fmt.Sprintf("monsters/%s", code),
		nil,
		&out,
	)

	if err != nil {
		return nil, err
	}

	CACHE_THAT_SHOULD_BE_DONE_BETTER_AND_ELSEWHERE[code] = &out

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
