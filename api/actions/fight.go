package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type FightResponse struct {
	Cooldown  types.Cooldown     `json:"cooldown"`
	Fight     types.FightDetails `json:"destination"`
	Character types.Character    `json:"character"`
}

func Fight(character string) (*FightResponse, error) {
	var payload = map[string]interface{}{}

	var out FightResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/fight", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
