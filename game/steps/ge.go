package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

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

func Sell(character string, code string, quantity_func QuantityCb, min_price int) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<ge/sell>", character))
	char, err := api.GetCharacterByName(character)
	if err != nil {
		log(fmt.Sprintf("failed to get character details for %s", character))
		return nil, err
	}

	item_details, err := api.GetGrandExchangeItemDetails(code)
	if err != nil {
		log(fmt.Sprintf("failed to get item details for %s", code))
		return nil, err
	}

	var quantity = quantity_func(CountInventory(&char.Inventory, code), item_details.Max_quantity)

	// Inventory Check
	// Quantity Calc Check
	if item_details.Max_quantity < quantity {
		log(fmt.Sprintf("can only sell %d %s, adjusting sell quantity ", quantity, code))
		quantity = min(item_details.Max_quantity, quantity)
	}

	// Price Check
	if item_details.Sell_price < min_price {
		log(fmt.Sprintf("%s is selling below min_price, sell_price %d < %d min_price", code, item_details.Sell_price, min_price))
		return nil, err
	}

	var price = max(item_details.Sell_price, min_price)

	log(fmt.Sprintf("selling %d %s for %d gp", quantity, code, price))
	res, err := actions.SellUnsafe(character, code, quantity, price)
	if err != nil {
		log(fmt.Sprintf("failed to sell %s", code))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res.Transaction))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func AutoSell(character string, code string, quantity int) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<ge/auto-sell>", character))
	char, err := api.GetCharacterByName(character)
	if err != nil {
		log(fmt.Sprintf("failed to get character details for %s", character))
		return nil, err
	}

	item_details, err := api.GetGrandExchangeItemDetails(code)
	if err != nil {
		log(fmt.Sprintf("failed to get item details for %s", code))
		return nil, err
	}

	// Inventory Check
	// Quantity Calc Check
	if item_details.Max_quantity < quantity {
		log(fmt.Sprintf("can only sell %d %s, adjusting sell quantity ", quantity, code))
		quantity = min(item_details.Max_quantity, quantity)
	}

	var bankRetrieveQuantity = 0
	var heldQuantity = CountInventory(&char.Inventory, code)
	if heldQuantity < quantity {
		var bankFetchQuantity = quantity - heldQuantity
		log(fmt.Sprintf("attempting to fetch %d %s from bank", bankFetchQuantity, code))
		bankInfo, err := api.GetBankItems()
		if err != nil {
			log("failed to get bank info")
			return nil, err
		}

		bankQuantity := CountBank(bankInfo, code)
		bankRetrieveQuantity = min(bankQuantity, bankFetchQuantity)
		if bankRetrieveQuantity > 0 {
			log(fmt.Sprintf("retrieving %d %s from bank", bankRetrieveQuantity, code))
			_, err = actions.BankWithdraw(character, code, bankRetrieveQuantity)
			if err != nil {
				log("failed to retrieve from bank")
				return nil, err
			}
		}
	}

	var price = item_details.Sell_price
	var sellQuantity = min(quantity, heldQuantity+bankRetrieveQuantity)

	log(fmt.Sprintf("selling %d %s for %d gp", sellQuantity, code, price))
	res, err := actions.SellUnsafe(character, code, quantity, price)
	if err != nil {
		log(fmt.Sprintf("failed to sell %s", code))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res.Transaction))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func Buy(character string, code string, quantity int, max_price int) (*types.Character, error) {
	// Inventory check?

	// Price Check
	log := utils.LogPre(fmt.Sprintf("[%s]<ge/buy>", character))
	item_details, err := api.GetGrandExchangeItemDetails(code)
	if err != nil {
		log(fmt.Sprintf("failed to get item details for %s", code))
		return nil, err
	}

	if max_price > 0 && item_details.Buy_price > max_price {
		log(fmt.Sprintf("%s is buying above max_price, buy_price %d > %d max_price", code, item_details.Buy_price, max_price))
		return nil, err
	}

	if item_details.Max_quantity < quantity {
		log(fmt.Sprintf("can only buy %d %s, adjusting buy quantity ", quantity, code))
		quantity = min(item_details.Max_quantity, quantity)
	}

	var price = item_details.Buy_price
	if max_price > 0 {
		price = min(item_details.Buy_price, max_price)
	}

	log(fmt.Sprintf("buying %d %s for %d gp", quantity, code, price))
	res, err := actions.BuyUnsafe(character, code, quantity, price)
	if err != nil {
		log(fmt.Sprintf("failed to buy %s", code))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res.Transaction))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
