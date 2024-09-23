package api

import (
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

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
