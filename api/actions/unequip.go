package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"github.com/mitchellh/mapstructure"
)

func UnequipItem(character string, slot string, quantity int) (*EquipResponse, error) {
	var payload = map[string]interface{}{
		"slot":     slot,
		"quantity": quantity,
	}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/unequip", character),
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
