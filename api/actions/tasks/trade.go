package tasks

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type TradeTaskItemResponse struct {
	Cooldown  types.Cooldown
	Trade     types.InventoryItem
	Character types.Character
}

func TradeTaskItem(character string, code string, quantity int) (*TradeTaskItemResponse, error) {
	payload := map[string]interface{}{
		"code":     code,
		"quantity": quantity,
	}

	var out TradeTaskItemResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/task/trade", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
