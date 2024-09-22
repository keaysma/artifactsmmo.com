package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"github.com/mitchellh/mapstructure"
)

type MoveResponse struct {
	Cooldown    api.Cooldown    `json:"cooldown"`
	Destination api.Destination `json:"destination"`
	Character   api.Character   `json:"character"`
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
