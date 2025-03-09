package steps

import (
	"fmt"

	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/utils"
)

func Gather(kernel *game.Kernel) error {
	log := kernel.LogPre(fmt.Sprintf("[%s]<gather>:", kernel.CharacterName))
	// Inventory check?

	log("Gathering")
	res, err := actions.Gather(kernel.CharacterName)
	if err != nil {
		log("Failed to gather")
		return err
	}

	kernel.DebugLog(utils.PrettyPrint(res.Details))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return nil
}
