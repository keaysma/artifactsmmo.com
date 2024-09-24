package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
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
	var payload = map[string]interface{}{}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/gathering", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out GatherResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
