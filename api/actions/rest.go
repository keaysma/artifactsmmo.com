package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type RestResponse struct {
	Cooldown    types.Cooldown
	Hp_restored int
	Character   types.Character
}

func Rest(character string) (*RestResponse, error) {
	var out RestResponse
	err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/rest", character),
		&map[string]interface{}{},
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
