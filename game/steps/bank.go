package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
)

type BankDepositCodeCb func(item types.InventorySlot) bool
type BankDepositQuantityCb func(item types.InventorySlot) int

func GetAllBankItems() (*[]types.InventoryItem, error) {
	page := 1
	allBankItems := make([]types.InventoryItem, 0)
	for {
		bankItems, err := api.GetBankItems(page)
		if err != nil {
			return nil, err
		}

		allBankItems = append(allBankItems, *bankItems...)

		if len(*bankItems) < api.GET_BANK_ITEMS_PAGE_SIZE {
			break
		}

		page++
	}

	return &allBankItems, nil
}

func SlotMaxQuantity() BankDepositQuantityCb {
	return func(item types.InventorySlot) int {
		return item.Quantity
	}
}

func DepositBySelect(kernel *game.Kernel, codeSelct BankDepositCodeCb, quantitySelect BankDepositQuantityCb) (*types.Character, error) {
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
			_, err := Move(character, coords.Bank)
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

	state.GlobalCharacter.With(func(value *types.Character) *types.Character {
		return char
	})

	return char, nil
}

func DepositEverything(kernel *game.Kernel) (*types.Character, error) {
	return DepositBySelect(character, func(slot types.InventorySlot) bool {
		return true
	}, SlotMaxQuantity())
}

type BankWithdrawCodeCb func(item types.InventoryItem) bool
type BankWithdrawQuantityCb func(item types.InventoryItem) int

func ItemMaxQuantity() BankWithdrawQuantityCb {
	return func(item types.InventoryItem) int {
		return item.Quantity
	}
}

func WithdrawBySelect(kernel *game.Kernel, codeSelect BankWithdrawCodeCb, quantitySelect BankWithdrawQuantityCb) (*types.Character, error) {
	var moved_to_bank = false

	bank, err := GetAllBankItems()
	if err != nil {
		return nil, err
	}

	var char *types.Character
	for _, slot := range *bank {
		if slot.Code == "" || !codeSelect(slot) {
			continue
		}

		quantity := quantitySelect(slot)
		if quantity == 0 {
			continue
		}

		// We have something to do, so go to the bank one time
		if !moved_to_bank {
			_, err := Move(character, coords.Bank)
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
		state.GlobalCharacter.Set(char)
		return char, nil
	}

	return nil, fmt.Errorf("no items to withdraw")
}

func DepositGold(kernel *game.Kernel, quantity int) (*types.Character, error) {
	_, err := Move(character, coords.Bank)
	if err != nil {
		return nil, err
	}

	res, err := actions.BankDepositGold(character, quantity)
	if err != nil {
		return nil, err
	}

	api.WaitForDown(res.Cooldown)
	state.GlobalCharacter.Set(&res.Character)

	return &res.Character, nil
}

func WithdrawGold(kernel *game.Kernel, quantity int) (*types.Character, error) {
	_, err := Move(character, coords.Bank)
	if err != nil {
		return nil, err
	}

	res, err := actions.BankWithdrawGold(character, quantity)
	if err != nil {
		return nil, err
	}

	api.WaitForDown(res.Cooldown)
	state.GlobalCharacter.Set(&res.Character)

	return &res.Character, nil
}
