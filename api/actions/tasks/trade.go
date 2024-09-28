package tasks

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
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

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/task/trade", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out TradeTaskItemResponse
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
