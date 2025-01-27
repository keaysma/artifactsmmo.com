package steps

import (
	"fmt"

	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func Rest(kernel *game.Kernel) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<rest>: ", kernel.CharacterName))
	log("Resting")
	res, err := actions.Rest(kernel.CharacterName)
	if err != nil {
		log(fmt.Sprintf("Failed to rest: %s", err))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res))
	log(fmt.Sprintf("Restored %d hp", res.Hp_restored))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
