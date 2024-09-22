package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"github.com/mitchellh/mapstructure"
)

type GatheringDetails struct {
	Xp    int
	Items []api.InventoryItem
}

type GatherResponse struct {
	Cooldown  api.Cooldown     `json:"cooldown"`
	Details   GatheringDetails `json:"destination"`
	Character api.Character    `json:"character"`
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
