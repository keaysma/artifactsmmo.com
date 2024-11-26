package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
)

func UnequipItem(character string, slot string, quantity int) (*EquipResponse, error) {
	var payload = map[string]interface{}{
		"slot":     slot,
		"quantity": quantity,
	}

	var out EquipResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/unequip", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
