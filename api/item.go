package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type ItemDetailResponse = types.ItemDetails

func GetItemDetails(code string) (*ItemDetailResponse, error) {
	utils.Log(fmt.Sprintf("Getting item details for %s", code))
	var out ItemDetailResponse
	err := GetDataResponseFuture(
		fmt.Sprintf("items/%s", code),
		nil,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
