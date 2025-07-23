package npc

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type NPCSellResponse struct {
	Cooldown    types.Cooldown
	Transaction types.Transaction
	Character   types.Character
}

func NPCSellItem(character string, code string, quantity int) (*NPCSellResponse, error) {
	payload := map[string]interface{}{
		"code":     code,
		"quantity": quantity,
	}

	var out NPCSellResponse
	err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/npc/sell", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
