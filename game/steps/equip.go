package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/utils"
)

func EquipItem(character string, code string, slot string, quantity int) error {
	char, err := api.GetCharacterByName(character)
	if err != nil {
		fmt.Printf("[%s][equip]: Failed to get character info\n", character)
		return err
	}

	cur_slot := utils.GetFieldFromStructByName(char, fmt.Sprintf("%s_slot", utils.Caser.String(slot))).String()
	if cur_slot != "" {
		fmt.Printf("[%s][equip]: Unequipping %s from %s\n", character, cur_slot, slot)
		err = UnequipItem(character, slot, 1)
		if err != nil {
			fmt.Printf("[%s][equip]: Failed to unequip %s\n", character, slot)
			return err
		}
	}

	fmt.Printf("[%s][equip]: Equipping %d %s to %s\n", character, quantity, code, slot)

	mres, err := actions.EquipItem(character, code, slot, quantity)
	if err != nil {
		fmt.Printf("[%s][equip]: Failed to equip %d %s to %s\n", character, quantity, code, slot)
		return err
	}

	if utils.GetSettings().Debug {
		fmt.Println(utils.PrettyPrint(mres.Item))
	}
	api.WaitForDown(mres.Cooldown)
	return nil
}

func UnequipItem(character string, slot string, quantity int) error {
	fmt.Printf("[%s][unequip]: Unequipping %d from %s\n", character, quantity, slot)

	mres, err := actions.UnequipItem(character, slot, quantity)
	if err != nil {
		fmt.Printf("[%s][unequip]: Failed to unequip %d from %s\n", character, quantity, slot)
		return err
	}

	if utils.GetSettings().Debug {
		fmt.Println(utils.PrettyPrint(mres.Item))
	}
	api.WaitForDown(mres.Cooldown)
	return nil
}
