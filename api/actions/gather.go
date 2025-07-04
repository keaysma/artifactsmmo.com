package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
)

type GatheringDetails struct {
	Xp    int
	Items []types.InventoryItem
}

type GatherResponse struct {
	Cooldown  types.Cooldown   `json:"cooldown"`
	Details   GatheringDetails `json:"destination"`
	Character types.Character  `json:"character"`
}

func Gather(character string) (*GatherResponse, error) {
	var out GatherResponse
	err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/gathering", character),
		nil,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
