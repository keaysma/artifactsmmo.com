package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/utils"
)

func EquipItem(character string, code string, slot string, quantity int) error {
	log := utils.LogPre(fmt.Sprintf("[%s]<equip>: ", character))
	char, err := api.GetCharacterByName(character)
	if err != nil {
		log("failed to get character info")
		return err
	}

	cur_slot := utils.GetFieldFromStructByName(char, fmt.Sprintf("%s_slot", utils.Caser.String(slot))).String()
	if cur_slot != "" {
		log(fmt.Sprintf("unequipping %d %s from %s", quantity, cur_slot, slot))
		err = UnequipItem(character, slot, 1)
		if err != nil {
			log(fmt.Sprintf("failed to unequip %s", slot))
			return err
		}
	}

	log(fmt.Sprintf("equipping %d %s to %s", quantity, code, slot))

	mres, err := actions.EquipItem(character, code, slot, quantity)
	if err != nil {
		log(fmt.Sprintf("failed to equip %d %s to %s", quantity, code, slot))
		return err
	}

	utils.DebugLog(utils.PrettyPrint(mres.Item))
	api.WaitForDown(mres.Cooldown)
	return nil
}

func UnequipItem(character string, slot string, quantity int) error {
	log := utils.LogPre(fmt.Sprintf("[%s]<unequip>: ", character))
	log(fmt.Sprintf("enequipping %d from %s", quantity, slot))

	mres, err := actions.UnequipItem(character, slot, quantity)
	if err != nil {
		log(fmt.Sprintf("failed to unequip %d from %s", quantity, slot))
		return err
	}

	utils.DebugLog(utils.PrettyPrint(mres.Item))
	api.WaitForDown(mres.Cooldown)
	return nil
}
