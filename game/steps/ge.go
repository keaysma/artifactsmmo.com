package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
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

func CountInventory(character *types.Character, code string) int {
	var total_quantity = 0
	if character == nil {
		return total_quantity
	}

	for s := range character.Inventory {
		if character.Inventory[s].Code == code {
			total_quantity += character.Inventory[s].Quantity
		}
	}

	return total_quantity
}

func Sell(character string, code string, quantity_func QuantityCb, min_price int) (*types.Character, error) {
	char, err := api.GetCharacterByName(character)
	if err != nil {
		fmt.Printf("[%s][ge/sell]: Failed to get character details for %s\n", character, character)
		return nil, err
	}

	item_details, err := actions.GetGrandExchangeItemDetails(code)
	if err != nil {
		fmt.Printf("[%s][ge/sell]: Failed to get item details for %s\n", character, code)
		return nil, err
	}

	var quantity = quantity_func(CountInventory(char, code), item_details.Max_quantity)

	// Inventory Check
	// Quantity Calc Check
	if item_details.Max_quantity < quantity {
		fmt.Printf("[%s][ge/sell]: Can only sell %d %s, adjusting sell quantity \n", character, quantity, code)
		quantity = min(item_details.Max_quantity, quantity)
	}

	// Price Check
	if item_details.Sell_price < min_price {
		fmt.Printf("[%s][ge/sell]: %s is selling below min_price, sell_price %d < %d min_price\n", character, code, item_details.Sell_price, min_price)
		return nil, err
	}

	var price = max(item_details.Sell_price, min_price)

	fmt.Printf("[%s][ge/sell]: Selling %d %s for %d gp\n", character, quantity, code, price)
	res, err := actions.SellUnsafe(character, code, quantity, price)
	if err != nil {
		fmt.Printf("[%s][ge/sell]: Failed to sell %s\n", character, code)
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res.Transaction))
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func Buy(character string, code string, quantity int, max_price int) (*types.Character, error) {
	// Inventory check?

	// Price Check
	item_details, err := actions.GetGrandExchangeItemDetails(code)
	if err != nil {
		fmt.Printf("[%s][ge/buy]: Failed to get item details for %s\n", character, code)
		return nil, err
	}

	if max_price > 0 && item_details.Buy_price > max_price {
		fmt.Printf("[%s][ge/buy]: %s is buying above max_price, buy_price %d > %d max_price\n", character, code, item_details.Buy_price, max_price)
		return nil, err
	}

	if item_details.Max_quantity < quantity {
		fmt.Printf("[%s][ge/buy]: Can only buy %d %s, adjusting buy quantity \n", character, quantity, code)
		quantity = min(item_details.Max_quantity, quantity)
	}

	var price = item_details.Buy_price
	if max_price > 0 {
		price = min(item_details.Buy_price, max_price)
	}

	fmt.Printf("[%s][ge/buy]: Buying %d %s for %d gp\n", character, quantity, code, price)
	res, err := actions.BuyUnsafe(character, code, quantity, price)
	if err != nil {
		fmt.Printf("[%s][ge/buy]: Failed to buy %s\n", character, code)
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res.Transaction))
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
