package utils

import "artifactsmmo.com/m/types"

type QuantityCb func(current_quantity int, max_quantity int) int

func Amount(amount int) QuantityCb { return func(_ int, _ int) int { return amount } }
func LeaveAtleast(amount int) QuantityCb {
	return func(current_quantity int, max_quantity int) int {
		if max_quantity >= (current_quantity + amount) {
			return current_quantity - amount
		}

		return max_quantity
	}
}

func CountInventory(slots *[]types.InventorySlot, code string) int {
	var total_quantity = 0
	if slots == nil {
		return total_quantity
	}

	for _, slot := range *slots {
		if slot.Code == code {
			total_quantity += slot.Quantity
		}
	}

	return total_quantity
}

func CountBank(slots *[]types.InventoryItem, code string) int {
	var total_quantity = 0
	if slots == nil {
		return total_quantity
	}

	for _, slot := range *slots {
		if slot.Code == code {
			total_quantity += slot.Quantity
		}
	}

	return total_quantity
}

func CountAllInventory(character *types.Character) int {
	var total_quantity = 0
	if character == nil {
		return total_quantity
	}

	for _, slot := range character.Inventory {
		total_quantity += slot.Quantity
	}

	return total_quantity
}

func CountSlotsInventory(character *types.Character) int {
	var total_slots = 0
	if character == nil {
		return total_slots
	}

	for _, slot := range character.Inventory {
		if slot.Quantity > 0 {
			total_slots++
		}
	}

	return total_slots
}

func FindInventorySlot(character *types.Character, code string) *types.InventorySlot {
	if character == nil {
		return nil
	}

	for _, slot := range character.Inventory {
		if slot.Code == code {
			return &types.InventorySlot{
				Slot:     slot.Slot,
				Quantity: slot.Quantity,
				Code:     slot.Code,
			}
		}
	}
	return nil
}
