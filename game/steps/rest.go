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
	log := utils.LogPre(fmt.Sprintf("[%s]<rest>: ", character))
	log("Resting")
	res, err := actions.Rest(character)
	if err != nil {
		log(fmt.Sprintf("Failed to rest: %s", err))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res))
	log(fmt.Sprintf("Restored %d hp", res.Hp_restored))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
