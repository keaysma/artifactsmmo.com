package tasks

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

type CompleteTaskResponse struct {
	Cooldown  types.Cooldown
	Reward    types.InventoryItem
	Character types.Character
}

func CompleteTask(character string) (*CompleteTaskResponse, error) {
	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/task/complete", character),
		&map[string]interface{}{},
	)

	if err != nil {
		return nil, err
	}

	var out CompleteTaskResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
