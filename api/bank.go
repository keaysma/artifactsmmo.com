package api

import (
	"artifactsmmo.com/m/types"
)

type BankDetailsResponse struct {
	Slots               int `json:"slots"`
	Expansions          int `json:"expansions"`
	Next_expansion_cost int `json:"next_expansion_cost"`
	Gold                int `json:"gold"`
}

func GetBankItems() (*[]types.InventoryItem, error) {
	var out []types.InventoryItem
	err := GetDataResponseFuture(
		"my/bank/items",
		&map[string]string{
			"size": "100",
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
	err := GetDataResponseFuture(
		"my/bank",
		nil,
		&out,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}
