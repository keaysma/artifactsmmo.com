package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
	"github.com/mitchellh/mapstructure"
)

type ItemCraftDetails struct {
	Skill    string
	Level    int
	Items    []types.InventoryItem
	Quantity int
}

type ItemDetails struct {
	Name        string
	Code        string
	Level       int
	Type        string
	Subtype     string
	Description string
	Effects     []types.Effect
	Craft       ItemCraftDetails
	Tradeable   bool
}

type ItemDetailResponse = ItemDetails

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
