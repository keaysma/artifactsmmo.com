package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func Gather(kernel *game.Kernel) (*types.Character, error) {
	// Inventory check?

	utils.Log(fmt.Sprintf("[%s]<gather>: Gathering ", character))
	res, err := actions.Gather(character)
	if err != nil {
		utils.Log(fmt.Sprintf("[%s]<gather>: Failed to gather", character))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res.Details))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
