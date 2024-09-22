package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/utils"
)

func Gather(character string) (*api.Character, error) {
	// Inventory check?

	fmt.Printf("[%s]: Gathering ", character)
	res, err := actions.Gather(character)
	if err != nil {
		fmt.Printf("[%s][gather]: Failed to gather", character)
		return nil, err
	}

	fmt.Println(utils.PrettyPrint(res.Details))
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
