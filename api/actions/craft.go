package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type CraftingDetails struct {
	Xp    int
	Items []types.InventoryItem
}

type CraftingResponse struct {
	Cooldown  types.Cooldown  `json:"cooldown"`
	Details   CraftingDetails `json:"details"`
	Character types.Character `json:"character"`
}

func Craft(character string, code string, quantity int) (*CraftingResponse, error) {
	var payload = map[string]interface{}{
		"code":     code,
		"quantity": quantity,
	}

	var out CraftingResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/crafting", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
