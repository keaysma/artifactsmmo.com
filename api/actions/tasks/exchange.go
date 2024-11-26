package tasks

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type ExchangeTaskCoinsResponse struct {
	Cooldown  types.Cooldown
	Reward    types.InventoryItem
	Character types.Character
}

func ExchangeTaskCoins(character string) (*ExchangeTaskCoinsResponse, error) {
	var out ExchangeTaskCoinsResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/task/exchange", character),
		&map[string]interface{}{},
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
