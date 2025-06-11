package api

import (
	"fmt"

	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

type ItemsResponse = []types.ItemDetails

const GET_ALL_ITEMS_PAGE_SIZE = 100

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

type GetAllItemsFilter struct {
	Itype          string
	Craft_material string
	Craft_skill    string
	Min_level      string
	Max_level      string
}

func GetAllItemsFiltered(filter GetAllItemsFilter, page int, size int) (*ItemsResponse, error) {
	utils.UniversalDebugLog(fmt.Sprintf("Getting all items with filter %v", filter))
	payload := map[string]string{
		"page": fmt.Sprintf("%d", page),
		"size": fmt.Sprintf("%d", size),
	}

	if filter.Itype != "" {
		payload["type"] = filter.Itype
	}

	if filter.Craft_material != "" {
		payload["craft_material"] = filter.Craft_material
	}

	if filter.Craft_skill != "" {
		payload["craft_skill"] = filter.Craft_skill
	}

	if filter.Min_level != "" {
		payload["min_level"] = filter.Min_level
	}

	if filter.Max_level != "" {
		payload["max_level"] = filter.Max_level
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
