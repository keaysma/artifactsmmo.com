package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/utils"
)

func EquipItem(kernel *game.Kernel, code string, slot string, quantity int) error {
	log := utils.LogPre(fmt.Sprintf("[%s]<equip>: ", character))
	char, err := api.GetCharacterByName(character)
	if err != nil {
		log("failed to get character info")
		return err
	}

	selectedSlot := slot
	if slot == "" {
		// automatically select slot
		itemDetails, err := api.GetItemDetails(code)
		if err != nil {
			log(fmt.Sprintf("failed to get item details for %s", code))
			return err
		}

		selectedSlot = itemDetails.Type

		// Special case for utility, rings, and artifacts
		// TODO: Select the best slot for these items based on the character's current equipment
		switch itemDetails.Type {
		case "utility", "ring", "artifact":
			selectedSlot = selectedSlot + "1"
		}
	}

	curSlot := utils.GetFieldFromStructByName(char, fmt.Sprintf("%s_slot", utils.Caser.String(selectedSlot))).String()
	if curSlot != "" {
		log(fmt.Sprintf("unequipping %d %s from %s", quantity, curSlot, selectedSlot))
		err = UnequipItem(character, selectedSlot, 1)
		if err != nil {
			log(fmt.Sprintf("failed to unequip %s", selectedSlot))
			return err
		}
	}

	log(fmt.Sprintf("equipping %d %s to %s", quantity, code, selectedSlot))

	mres, err := actions.EquipItem(character, code, selectedSlot, quantity)
	if err != nil {
		log(fmt.Sprintf("failed to equip %d %s to %s", quantity, code, selectedSlot))
		return err
	}

	utils.DebugLog(utils.PrettyPrint(mres.Item))
	api.WaitForDown(mres.Cooldown)
	return nil
}

func UnequipItem(kernel *game.Kernel, slot string, quantity int) error {
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
