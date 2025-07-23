package api

import "fmt"

type NPCDetails struct {
	Name        string
	Code        string
	Description string
	Type        string
}

type NPCItem struct {
	Code       string
	Npc        string
	Currency   string
	Buy_price  *int
	Sell_price *int
}

func GetAllNPCs() (*[]NPCDetails, error) {
	var out []NPCDetails
	err := GetDataResponse(
		"npcs/details",
		nil,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

func GetNPCByCode(code string) (*NPCDetails, error) {
	var out NPCDetails
	err := GetDataResponse(
		fmt.Sprintf("npcs/details/%s", code),
		nil,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

type GetAllNPCItemsParams struct {
	Code *string
	Npc  *string
	Page *int
	Size *int
}

func GetAllNPCItems(in GetAllNPCItemsParams) (*[]NPCItem, error) {
	payload := map[string]string{}

	if in.Code != nil {
		payload["code"] = *in.Code
	}

	if in.Npc != nil {
		payload["npc"] = *in.Npc
	}

	if in.Page != nil {
		payload["page"] = fmt.Sprintf("%d", *in.Page)
	}

	if in.Size != nil {
		payload["size"] = fmt.Sprintf("%d", *in.Size)
	}

	var out []NPCItem
	err := GetDataResponse(
		"npcs/items",
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}

func GetNPCItems(code string, page int, size int) (*[]NPCItem, error) {
	payload := map[string]string{
		"page": fmt.Sprintf("%d", page),
		"size": fmt.Sprintf("%d", size),
	}

	var out []NPCItem
	err := GetDataResponse(
		fmt.Sprintf("/npcs/items/%s", code),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
