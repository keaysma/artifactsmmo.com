package tasks

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

type ExchangeTaskCoinsResponse struct {
	Cooldown  types.Cooldown
	Reward    types.InventoryItem
	Character types.Character
}

func ExchangeTaskCoins(character string) (*ExchangeTaskCoinsResponse, error) {
	res, err := api.PostDataResponse(
		fmt.Sprintf("/my/%s/action/task/exchange", character),
		&map[string]interface{}{},
	)

	if err != nil {
		return nil, err
	}

	var out ExchangeTaskCoinsResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
