package tasks

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type NewTaskResponse struct {
	Cooldown  types.Cooldown
	Task      types.Task
	Character types.Character
}

func NewTask(character string) (*NewTaskResponse, error) {
	var out NewTaskResponse
	err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/task/new", character),
		&map[string]interface{}{},
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
