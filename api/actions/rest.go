package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

type RestResponse struct {
	Cooldown    types.Cooldown
	Hp_restored int
	Character   types.Character
}

func Rest(character string) (*RestResponse, error) {
	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/rest", character),
		&map[string]interface{}{},
	)

	if err != nil {
		return nil, err
	}

	var out RestResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
