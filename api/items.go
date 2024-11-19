package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
	"github.com/mitchellh/mapstructure"
)

type ItemsResponse = []types.ItemDetails

func GetAllItemsByType(item_type string, page int, size int) (*ItemsResponse, error) {
	utils.Log(fmt.Sprintf("Getting all items of type %s", item_type))
	payload := map[string]string{
		"type": item_type,
		"page": fmt.Sprintf("%d", page),
		"size": fmt.Sprintf("%d", size),
	}

	res, err := GetDataResponse(
		"items",
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out ItemsResponse
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}

func GetAllItemsByCraftSkill(craft_skill string, page int, size int) (*ItemsResponse, error) {
	utils.Log(fmt.Sprintf("Getting all items of craft_skill %s", craft_skill))
	payload := map[string]string{
		"craft_skill": craft_skill,
		"page":        fmt.Sprintf("%d", page),
		"size":        fmt.Sprintf("%d", size),
	}

	res, err := GetDataResponse(
		"items",
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out ItemsResponse
	uerr := mapstructure.Decode(res.Data, &out)
	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
