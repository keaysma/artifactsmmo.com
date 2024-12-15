package tasks

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type CancelTaskResponse struct {
	Cooldown  types.Cooldown
	Character types.Character
}

func CancelTask(character string) (*CompleteTaskResponse, error) {
	var out CompleteTaskResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/task/cancel", character),
		&map[string]interface{}{},
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
