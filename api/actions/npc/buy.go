package npc

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type NPCBuyResponse struct {
	Cooldown    types.Cooldown
	Transaction types.Transaction
	Character   types.Character
}

func NPCBuyItem(character string, code string, quantity int) (*NPCBuyResponse, error) {
	payload := map[string]interface{}{
		"code":     code,
		"quantity": quantity,
	}

	var out NPCBuyResponse
	err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/npc/buy", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
