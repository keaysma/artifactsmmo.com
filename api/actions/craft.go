package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
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

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/crafting", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out CraftingResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
