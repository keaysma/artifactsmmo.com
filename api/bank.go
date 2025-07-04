package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
)

type BankDetailsResponse struct {
	Slots               int `json:"slots"`
	Expansions          int `json:"expansions"`
	Next_expansion_cost int `json:"next_expansion_cost"`
	Gold                int `json:"gold"`
}

func GetBankItemByCode(code string) (*[]types.InventoryItem, error) {
	var out []types.InventoryItem
	err := GetDataResponse(
		"my/bank/items",
		&map[string]string{
			"code": code,
			"size": "1",
		},
		&out,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

const GET_BANK_ITEMS_PAGE_SIZE = 100

func GetBankItems(page int) (*[]types.InventoryItem, error) {
	var out []types.InventoryItem
	err := GetDataResponse(
		"my/bank/items",
		&map[string]string{
			"page": fmt.Sprintf("%d", page),
			"size": fmt.Sprintf("%d", GET_BANK_ITEMS_PAGE_SIZE),
		},
		&out,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func GetBankDetails() (*BankDetailsResponse, error) {
	var out BankDetailsResponse
	err := GetDataResponse(
		"my/bank",
		nil,
		&out,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}
