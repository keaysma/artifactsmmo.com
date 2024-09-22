package api

import (
	"github.com/mitchellh/mapstructure"
)

func GetBankItems() (*[]InventoryItem, error) {
	res, err := GetDataResponse("my/bank/items", nil)
	if err != nil {
		return nil, err
	}

	var out []InventoryItem
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
