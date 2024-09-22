package strategies

import (
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game/steps"
)

// Buy and Equip is a pretty common workflow to automate

func BuyAndEquip(character string, code string, price int, slot string) error {
	err := steps.Move(character, coords.GrandExchange)
	if err != nil {
		return err
	}

	_, buy_err := steps.Buy(character, code, 1, price)
	if buy_err != nil {
		return buy_err
	}

	uneq_err := steps.UnequipItem(character, slot, 1)
	if uneq_err != nil {
		return uneq_err
	}

	eq_err := steps.EquipItem(character, code, slot, 1)
	if eq_err != nil {
		return eq_err
	}

	return nil
}
