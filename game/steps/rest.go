package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func Rest(character string) (*types.Character, error) {
	utils.Log(fmt.Sprintf("[%s]<rest>: Resting", character))
	res, err := actions.Rest(character)
	if err != nil {
		utils.Log(fmt.Sprintf("[%s]<rest>: Failed to rest", character))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res))
	utils.Log(fmt.Sprintf("[%s]<resr>: Restored %d hp", character, res.Hp_restored))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
