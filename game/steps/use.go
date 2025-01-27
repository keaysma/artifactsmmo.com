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

func Use(kernel *game.Kernel, code string, quantity int) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<use>: ", character))
	res, err := actions.Use(character, code, quantity)
	if err != nil {
		log(fmt.Sprintf("Failed to use %d %s: %s", quantity, code, err))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res))
	log(fmt.Sprintf("Used %d %s", quantity, code))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
