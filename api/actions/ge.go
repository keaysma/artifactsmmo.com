package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"github.com/mitchellh/mapstructure"
)

type TransactionDetails struct {
	Code        string
	Quantity    int
	Price       int
	Total_price int
}

type TransactionResponse struct {
	Cooldown    api.Cooldown       `json:"cooldown"`
	Transaction TransactionDetails `json:"transaction"`
	Character   api.Character      `json:"character"`
}

func SellUnsafe(character string, code string, quantity int, price int) (*TransactionResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"quantity": quantity,
		"price":    price,
	}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/ge/sell", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out TransactionResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}

func BuyUnsafe(character string, code string, quantity int, price int) (*TransactionResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"quantity": quantity,
		"price":    price,
	}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/ge/buy", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out TransactionResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}

func GetGrandExchangeItemDetails(code string) (*api.GrandExchangeItemData, error) {
	res, err := api.GetDataResponse(
		fmt.Sprintf("ge/%s", code),
		nil,
	)

	if err != nil {
		return nil, err
	}

	var out api.GrandExchangeItemData
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
