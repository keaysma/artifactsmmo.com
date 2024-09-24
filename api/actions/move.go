package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

type MoveResponse struct {
	Cooldown    types.Cooldown    `json:"cooldown"`
	Destination types.Destination `json:"destination"`
	Character   types.Character   `json:"character"`
}

func Move(character string, x int, y int) (*MoveResponse, error) {
	var payload = map[string]interface{}{
		"x": x,
		"y": y,
	}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/move", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out MoveResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
