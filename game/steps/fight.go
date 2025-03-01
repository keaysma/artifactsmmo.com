package steps

import (
	"fmt"

	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/utils"
)

func FightUnsafe(kernel *game.Kernel, print_fight_logs bool) error {
	log := utils.DebugLogPre(fmt.Sprintf("[%s]<fight>: ", kernel.CharacterName))
	log("Fighting (unsafe call)!")

	mres, err := actions.Fight(kernel.CharacterName)
	if err != nil {
		log("Failed to fight")
		return err
	}

	custom_details := map[string]interface{}{
		"result": mres.Fight.Result,
		"xp":     mres.Fight.Xp,
		"gold":   mres.Fight.Gold,
		"drops":  mres.Fight.Drops,
		"hp":     mres.Character.Hp,
	}
	if print_fight_logs {
		for _, log := range mres.Fight.Logs {
			utils.Log(log)
		}
		utils.Log(fmt.Sprintln(utils.PrettyPrint(custom_details)))
	} else {
		utils.DebugLog(fmt.Sprintln(utils.PrettyPrint(custom_details)))
	}
	log(fmt.Sprintf("Result: %s", mres.Fight.Result))

	kernel.WaitForDown(mres.Cooldown)
	kernel.CharacterState.Set(&mres.Character)

	if mres.Fight.Result != "win" {
		return fmt.Errorf("[%s]<fight>: result is %s", kernel.CharacterName, mres.Fight.Result)
	}

	return nil
}

var HP_SAFETY_PERCENT = 0.51

func FightSafeHpAmount(max_hp int) int {
	return int(float64(max_hp) * HP_SAFETY_PERCENT)
}

func FightHpSafetyCheck(hp int, max_hp int) bool {
	hpSafety := FightSafeHpAmount(max_hp)
	return hp >= hpSafety
}

func Fight(kernel *game.Kernel) error {
	log := utils.LogPre(fmt.Sprintf("[%s]<fight>: ", kernel.CharacterName))
	log("Fighting")

	character := kernel.CharacterState.Ref()
	hp, max_hp := character.Hp, character.Max_hp
	kernel.CharacterState.Unlock()

	if !FightHpSafetyCheck(hp, max_hp) {
		log(fmt.Sprintf("Will not fight, HP below safety (%d < %d)", hp, FightSafeHpAmount(max_hp)))
		return nil
	}

	return FightUnsafe(kernel, false)
}

func FightDebug(kernel *game.Kernel) error {
	return FightUnsafe(kernel, true)
}
