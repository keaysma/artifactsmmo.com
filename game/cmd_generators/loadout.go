package generators

import (
	"fmt"

	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/loadout"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

var LoadOutForFight = loadout.LoadOutForFightV2

func LoadoutCommandFromResults(kernel *game.Kernel, loadout map[string]*types.ItemDetails) string {
	if len(loadout) == 0 {
		return ""
	}

	cmd := "loadout "
	for slot, item := range loadout {
		char := kernel.CharacterState.Ref()
		curItem := utils.GetFieldFromStructByName(char, fmt.Sprintf("%s_slot", slot)).String()
		kernel.CharacterState.Unlock()

		if curItem == item.Code {
			continue
		}

		cmd += fmt.Sprintf("%s:%s ", slot, item.Code)
	}

	// snip that extra space
	cmd = cmd[:len(cmd)-1]

	return cmd
}

func LoadOutCommand(kernel *game.Kernel, target string) (string, error) {
	hp, maxHp := 0, 0
	kernel.CharacterState.Read(func(value *types.Character) {
		hp = value.Hp
		maxHp = value.Max_hp
	})

	if hp < maxHp {
		return "rest", nil
	}

	loadout, err := LoadOutForFight(kernel, target)
	if err != nil {
		return "", err
	}

	return LoadoutCommandFromResults(kernel, loadout), nil
}
