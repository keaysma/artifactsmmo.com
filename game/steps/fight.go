package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func FightUnsafe(character string) (*types.Character, error) {
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

var HP_SAFETY_PERCENT = 0.51

func FightSafeHpAmount(max_hp int) int {
	return int(float64(max_hp) * HP_SAFETY_PERCENT)
}

func FightHpSafetyCheck(hp int, max_hp int) bool {
	hpSafety := FightSafeHpAmount(max_hp)
	return hp >= hpSafety
}

func Fight(character string) (*types.Character, error) {
	utils.Log(fmt.Sprintf("[%s]<fight>: Fighting", character))

	char_start, err := api.GetCharacterByName(character)
	if err != nil || char_start == nil {
		utils.Log(fmt.Sprintf("[%s]<fight>: Failed to get character info", character))
		return nil, err
	}

	state.GlobalCharacter.With(func(value *types.Character) *types.Character {
		return char_start
	})

	if !FightHpSafetyCheck(char_start.Hp, char_start.Max_hp) {
		utils.Log(fmt.Sprintf("[%s]<fight>: Will not fight, HP below safety (%d < %d)", character, char_start.Hp, FightSafeHpAmount(char_start.Max_hp)))
		return char_start, nil
	}

	char_end, err := FightUnsafe(character)
	if err != nil {
		utils.Log(fmt.Sprintf("[%s]<fight>: Failed to fight", character))
		return nil, err
	}

	state.GlobalCharacter.With(func(value *types.Character) *types.Character {
		return char_end
	})

	return char_end, nil
}
