package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func Gather(character string) (*types.Character, error) {
	// Inventory check?

	utils.Log(fmt.Sprintf("[%s]<gather>: Gathering ", character))
	res, err := actions.Gather(character)
	if err != nil {
		utils.Log(fmt.Sprintf("[%s]<gather>: Failed to gather", character))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res.Details))
	state.GlobalCharacter.With(func(value *types.Character) *types.Character {
		return &res.Character
	})
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
