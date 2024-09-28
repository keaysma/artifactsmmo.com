package tasks

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

type NewTaskResponse struct {
	Cooldown  types.Cooldown
	Task      types.Task
	Character types.Character
}

func NewTask(character string) (*NewTaskResponse, error) {
	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/task/new", character),
		&map[string]interface{}{},
	)

	if err != nil {
		return nil, err
	}

	var out NewTaskResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
