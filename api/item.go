package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
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
}

type ItemDetailResponse struct {
	Item ItemDetails                 `json:"item"`
	Ge   types.GrandExchangeItemData `json:"ge"`
}

func GetItemDetails(code string) (*ItemDetailResponse, error) {
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
