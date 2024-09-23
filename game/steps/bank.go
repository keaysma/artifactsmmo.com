package steps

import (
	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/types"
)

type BankDepositCodeCb func(item api.InventorySlot) bool
type BankDepositQuantityCb func(item api.InventorySlot) int

func SlotMaxQuantity() BankDepositQuantityCb {
	return func(item api.InventorySlot) int {
		return item.Quantity
	}
}

func DepositBySelect(character string, codeSelct BankDepositCodeCb, quantitySelect BankDepositQuantityCb) (*api.Character, error) {
	var moved_to_bank = false

	char, err := api.GetCharacterByName(character)
	if err != nil {
		return nil, err
	}

	for _, slot := range char.Inventory {
		if slot.Code == "" || !codeSelct(slot) {
			continue
		}

		quantity := quantitySelect(slot)
		if quantity == 0 {
			continue
		}

		// We have something to do, so go to the bank one time
		if !moved_to_bank {
			err := Move(character, coords.Bank)
			if err != nil {
				return nil, err
			}
			moved_to_bank = true
		}

		res, err := actions.BankDeposit(character, slot.Code, quantity)
		if err != nil {
			return nil, err
		}

		char = &res.Character
		api.WaitForDown(res.Cooldown)
	}

	return char, nil
}

type BankWithdrawCodeCb func(item types.InventoryItem) bool
type BankWithdrawQuantityCb func(item types.InventoryItem) int

func ItemMaxQuantity() BankWithdrawQuantityCb {
	return func(item types.InventoryItem) int {
		return item.Quantity
	}
}

func WithdrawBySelect(character string, codeSelct BankWithdrawCodeCb, quantitySelect BankWithdrawQuantityCb) (*api.Character, error) {
	var moved_to_bank = false

	bank, err := api.GetBankItems()
	if err != nil {
		return nil, err
	}

	var char *api.Character
	for _, slot := range *bank {
		if slot.Code == "" || !codeSelct(slot) {
			continue
		}

		quantity := quantitySelect(slot)
		if quantity == 0 {
			continue
		}

		// We have something to do, so go to the bank one time
		if !moved_to_bank {
			err := Move(character, coords.Bank)
			if err != nil {
				return nil, err
			}
			moved_to_bank = true
		}

		res, err := actions.BankWithdraw(character, slot.Code, quantity)
		if err != nil {
			return nil, err
		}

		char = &res.Character
		api.WaitForDown(res.Cooldown)
	}

	return char, nil
}
