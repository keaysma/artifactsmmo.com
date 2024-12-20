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

type BankGoldResponse struct {
	Cooldown types.Cooldown
	Bank     struct {
		Quantity int
	}
	Character types.Character
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

func BankDepositGold(character string, quantity int) (*BankGoldResponse, error) {
	var payload = map[string]interface{}{
		"quantity": quantity,
	}

	var out BankGoldResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/bank/deposit/gold", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

func BankWithdrawGold(character string, quantity int) (*BankGoldResponse, error) {
	var payload = map[string]interface{}{
		"quantity": quantity,
	}

	var out BankGoldResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/bank/withdraw/gold", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
