package steps

import (
	"fmt"

	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/utils"
)

func Use(kernel *game.Kernel, code string, quantity int) error {
	log := utils.LogPre(fmt.Sprintf("[%s]<use>: ", kernel.CharacterName))
	res, err := actions.Use(kernel.CharacterName, code, quantity)
	if err != nil {
		log(fmt.Sprintf("Failed to use %d %s: %s", quantity, code, err))
		return err
	}

	utils.DebugLog(utils.PrettyPrint(res))
	log(fmt.Sprintf("Used %d %s", quantity, code))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return nil
}
