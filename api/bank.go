package api

import (
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

type BankDetailsResponse struct {
	Slots               int `json:"slots"`
	Expansions          int `json:"expansions"`
	Next_expansion_cost int `json:"next_expansion_cost"`
	Gold                int `json:"gold"`
}

func GetBankItems() (*[]types.InventoryItem, error) {
	res, err := GetDataResponse("my/bank/items", nil)
	if err != nil {
		return nil, err
	}

	var out []types.InventoryItem
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}

func GetBankDetails() (*BankDetailsResponse, error) {
	res, err := GetDataResponse("my/bank", nil)
	if err != nil {
		return nil, err
	}

	var out BankDetailsResponse
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
