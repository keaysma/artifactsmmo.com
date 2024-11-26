package actions

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/types"
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

	var out FightResponse
	err := api.PostDataResponseFuture(
		fmt.Sprintf("my/%s/action/fight", character),
		&payload,
		&out,
	)

	if err != nil {
		return nil, err
	}

	return &out, nil
}
