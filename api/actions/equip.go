package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type EquipResponse struct {
	Cooldown  types.Cooldown    `json:"cooldown"`
	Slot      string            `json:"slot"`
	Item      types.ItemDetails `json:"item"`
	Character types.Character   `json:"character"`
}

func EquipItem(character string, code string, slot string, quantity int) (*EquipResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"slot":     slot,
		"quantity": quantity,
	}

	var out EquipResponse
	err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/equip", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
