package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
	"github.com/mitchellh/mapstructure"
)

type BlockedHits struct {
	Fire  int
	Earth int
	Water int
	Air   int
	Total int
}

type FightDetails struct {
	Xp                   int
	Gold                 int
	Drops                []types.InventoryItem
	Turns                int
	Monster_blocked_hits BlockedHits
	Player_blocked_hits  BlockedHits
	Logs                 []string
	Result               string
}

type FightResponse struct {
	Cooldown  types.Cooldown  `json:"cooldown"`
	Fight     FightDetails    `json:"destination"`
	Character types.Character `json:"character"`
}

func Fight(character string) (*FightResponse, error) {
	var payload = map[string]interface{}{}

	res, err := api.PostDataResponse(
		fmt.Sprintf("my/%s/action/fight", character),
		&payload,
	)

	if err != nil {
		return nil, err
	}

	var out FightResponse
	uerr := mapstructure.Decode(res.Data, &out)

	if uerr != nil {
		return nil, uerr
	}

	return &out, nil
}
