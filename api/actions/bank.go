package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type BankItemResponse struct {
	Cooldown  types.Cooldown        `json:"cooldown"`
	Item      types.ItemDetails     `json:"item"`
	Bank      []types.InventoryItem `json:"bank"`
	Character types.Character       `json:"character"`
}

func BankDeposit(character string, code string, quantity int) (*BankItemResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"quantity": quantity,
	}

	var out BankItemResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/bank/deposit", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

func BankWithdraw(character string, code string, quantity int) (*BankItemResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"quantity": quantity,
	}

	var out BankItemResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/bank/withdraw", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
