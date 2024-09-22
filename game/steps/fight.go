package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/utils"
)

func FightUnsafe(character string) (*api.Character, error) {
	fmt.Printf("[%s][fight]: Fighting!\n", character)

	mres, err := actions.Fight(character)
	if err != nil {
		fmt.Printf("[%s][fight]: Failed to fight\n", character)
		return nil, err
	}

	custom_details := map[string]interface{}{
		"result": mres.Fight.Result,
		"xp":     mres.Fight.Xp,
		"gold":   mres.Fight.Gold,
		"drops":  mres.Fight.Drops,
		"hp":     mres.Character.Hp,
	}
	fmt.Println(utils.PrettyPrint(custom_details))

	api.WaitForDown(mres.Cooldown)

	if mres.Fight.Result != "win" {
		fmt.Printf("[%s][fight]: Result is not win: %s\n", character, mres.Fight.Result)
		return nil, fmt.Errorf("[%s][fight]: result is %s", character, mres.Fight.Result)
	}

	return &mres.Character, nil
}

func Fight(character string, hpSafety int) (*api.Character, error) {
	fmt.Printf("[%s][fight]: Fighting!\n", character)

	char_start, err := api.GetCharacterByName(character)
	if err != nil {
		fmt.Printf("[%s][fight]: Failed to get character info\n", character)
		return nil, err
	}

	if char_start.Hp < hpSafety {
		fmt.Printf("[%s][fight]: Will not fight, HP below safety (%d < %d)\n", character, char_start.Hp, hpSafety)
		return char_start, nil
	}

	char_end, err := FightUnsafe(character)
	if err != nil {
		fmt.Printf("[%s][fight]: Failed to fight\n", character)
		return nil, err
	}

	return char_end, nil
}
