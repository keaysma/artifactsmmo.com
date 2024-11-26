package tasks

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type CompleteTaskResponse struct {
	Cooldown  types.Cooldown
	Reward    types.InventoryItem
	Character types.Character
}

func CompleteTask(character string) (*CompleteTaskResponse, error) {
	var out CompleteTaskResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/task/complete", character),
		&map[string]interface{}{},
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
