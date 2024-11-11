package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

type CreateSellOrderResponse struct {
	Cooldown  types.Cooldown
	Order     types.CreateOrderData
	Character types.Character
}

type CancelSellOrderResponse struct {
	Cooldown  types.Cooldown
	Order     types.Order
	Character types.Character
}

type BuyItemResponse struct {
	Cooldown  types.Cooldown
	Order     types.Order
	Character types.Character
}

func CreateSellOrder(character string, code string, quantity int, price int) (*CreateSellOrderResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"quantity": quantity,
		"price":    price,
	}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/grandexchange/sell", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out CreateSellOrderResponse
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}

func CancelSellOrder(character string, id string) (*CancelSellOrderResponse, error) {
	var payload = map[string]interface{}{
		"id": id,
	}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/grandexchange/cancel", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out CancelSellOrderResponse
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, err
	}

	return &out, nil
}

func HitSellOrder(character string, id string, quantity int) (*BuyItemResponse, error) {
	var payload = map[string]interface{}{
		"id":       id,
		"quantity": quantity,
	}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/grandexchange/buy", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out BuyItemResponse
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, err
	}

	return &out, nil
}
