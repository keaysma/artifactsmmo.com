package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

func GetGrandExchangeItemDetails(code string) (*types.GrandExchangeItemData, error) {
	res, err := GetDataResponse(
		fmt.Sprintf("ge/%s", code),
		nil,
	)

	if err != nil {
		return nil, err
	}

	var out types.GrandExchangeItemData
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}

func GetAllGrandExchangeItemDetails(page int, size int) (*[]types.GrandExchangeItemData, error) {
	payload := map[string]string{
		"page": fmt.Sprintf("%d", page),
		"size": fmt.Sprintf("%d", size),
	}
	res, err := GetDataResponse("ge", &payload)

	if err != nil {
		return nil, err
	}

	var out []types.GrandExchangeItemData
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
