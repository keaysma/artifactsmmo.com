package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"github.com/mitchellh/mapstructure"
)

type BankItemResponse struct {
	Cooldown  api.Cooldown        `json:"cooldown"`
	Item      api.ItemDetails     `json:"item"`
	Bank      []api.InventoryItem `json:"bank"`
	Character api.Character       `json:"character"`
}

func BankDeposit(character string, code string, quantity int) (*BankItemResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"quantity": quantity,
	}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/bank/deposit", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out BankItemResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}

func BankWithdraw(character string, code string, quantity int) (*BankItemResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"quantity": quantity,
	}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/bank/withdraw", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out BankItemResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
