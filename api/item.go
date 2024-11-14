package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
	"github.com/mitchellh/mapstructure"
)

type ItemDetailResponse = types.ItemDetails

func GetItemDetails(code string) (*ItemDetailResponse, error) {
	utils.Log(fmt.Sprintf("Getting item details for %s", code))
	res, err := GetDataResponse(
		fmt.Sprintf("items/%s", code),
		nil,
	)

	if err != nil {
		return nil, err
	}

	var out ItemDetailResponse
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
