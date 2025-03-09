package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type ItemsResponse = []types.ItemDetails

func GetAllItemsByType(item_type string, page int, size int) (*ItemsResponse, error) {
	utils.UniversalDebugLog(fmt.Sprintf("Getting all items of type %s", item_type))
	payload := map[string]string{
		"type": item_type,
		"page": fmt.Sprintf("%d", page),
		"size": fmt.Sprintf("%d", size),
	}

	var out ItemsResponse
	err := GetDataResponseFuture(
		"items",
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

func GetAllItemsByCraftSkill(craft_skill string, page int, size int) (*ItemsResponse, error) {
	utils.UniversalDebugLog(fmt.Sprintf("Getting all items of craft_skill %s", craft_skill))
	payload := map[string]string{
		"craft_skill": craft_skill,
		"page":        fmt.Sprintf("%d", page),
		"size":        fmt.Sprintf("%d", size),
	}

	var out ItemsResponse
	err := GetDataResponseFuture(
		"items",
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
