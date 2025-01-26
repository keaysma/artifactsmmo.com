package tasks

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type GetTaskResponse struct {
	Code         string
	Level        int
	Type         string
	Min_quantity int
	Max_quantity int
	Skill        string
	Rewards      struct {
		Items []types.InventoryItem
		Gold  int
	}
}

func GetTask(code string) (*GetTaskResponse, error) {
	var out GetTaskResponse
	err := api.GetDataResponseFuture(
		fmt.Sprintf("tasks/list/%s", code),
		nil,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
