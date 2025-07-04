package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
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

	var out MoveResponse
	err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/move", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
