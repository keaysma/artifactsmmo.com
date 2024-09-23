package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/utils"
)

func FightUnsafe(character string) (*api.Character, error) {
	utils.DebugLog(fmt.Sprintf("[%s]<fight>: Fighting (unsafe call)!", character))

	mres, err := actions.Fight(character)
	if err != nil {
		utils.DebugLog(fmt.Sprintf("[%s]<fight>: Failed to fight", character))
		return nil, err
	}

	custom_details := map[string]interface{}{
		"result": mres.Fight.Result,
		"xp":     mres.Fight.Xp,
		"gold":   mres.Fight.Gold,
		"drops":  mres.Fight.Drops,
		"hp":     mres.Character.Hp,
	}
	utils.DebugLog(fmt.Sprintln(utils.PrettyPrint(custom_details)))

	api.WaitForDown(mres.Cooldown)

	if mres.Fight.Result != "win" {
		utils.Log(fmt.Sprintf("[%s]<fight>: Result is not win: %s\n", character, mres.Fight.Result))
		return nil, fmt.Errorf("[%s]<fight>: result is %s", character, mres.Fight.Result)
	}

	return &mres.Character, nil
}

func Fight(character string, hpSafety int) (*api.Character, error) {
	utils.Log(fmt.Sprintf("[%s]<fight>: Fighting", character))

	char_start, err := api.GetCharacterByName(character)
	if err != nil || char_start == nil {
		utils.Log(fmt.Sprintf("[%s]<fight>: Failed to get character info", character))
		return nil, err
	}

	if char_start.Hp < hpSafety {
		utils.Log(fmt.Sprintf("[%s]<fight>: Will not fight, HP below safety (%d < %d)", character, char_start.Hp, hpSafety))
		return char_start, nil
	}

	char_end, err := FightUnsafe(character)
	if err != nil {
		utils.Log(fmt.Sprintf("[%s]<fight>: Failed to fight", character))
		return nil, err
	}

	return char_end, nil
}
