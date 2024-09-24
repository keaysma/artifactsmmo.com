package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

type EquipResponse struct {
	Cooldown  types.Cooldown  `json:"cooldown"`
	Slot      string          `json:"slot"`
	Item      api.ItemDetails `json:"item"`
	Character types.Character `json:"character"`
}

func EquipItem(character string, code string, slot string, quantity int) (*EquipResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"slot":     slot,
		"quantity": quantity,
	}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/equip", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out EquipResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
