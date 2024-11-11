package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func Sell(character string, code string, quantity int, minPrice int) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<ge/sell>", character))

	_, err := Move(character, coords.GrandExchange)
	if err != nil {
		return nil, err
	}

	sellOrders, err := api.GetSellOrders(code, nil)
	if err != nil {
		log(fmt.Sprintf("failed to get sell orders for %s: %s", code, err))
		return nil, err
	}

	sellOrderHistory, err := api.GetSellOrderHistory(code, nil, nil)
	if err != nil {
		log(fmt.Sprintf("failed to get sell order history for %s: %s", code, err))
		return nil, err
	}

	averageSellPrice := 0
	for _, order := range *sellOrders {
		averageSellPrice += order.Price
	}
	for _, order := range *sellOrderHistory {
		averageSellPrice += order.Price
	}

	averageSellPrice /= max(1, len(*sellOrders)+len(*sellOrderHistory))

	// Price Check
	if minPrice < averageSellPrice {
		log(fmt.Sprintf("min price %d for item %s is below the historical average: %d", minPrice, code, averageSellPrice))
		return nil, err
	}

	var price = max(averageSellPrice, minPrice)

	log(fmt.Sprintf("creating sell order for %d %s at %d gp", quantity, code, price))
	res, err := actions.CreateSellOrder(character, code, quantity, price)
	if err != nil {
		log(fmt.Sprintf("failed to sell %s", code))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res.Order))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func Buy(character string, code string, quantity int, maxPrice int) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<ge/buy>", character))

	_, err := Move(character, coords.GrandExchange)
	if err != nil {
		return nil, err
	}

	sellOrders, err := api.GetSellOrders(code, nil)
	if err != nil {
		log(fmt.Sprintf("failed to get sell orders for %s: %s", code, err))
		return nil, err
	}

	if len(*sellOrders) == 0 {
		log(fmt.Sprintf("no sell orders for %s", code))
		return nil, err
	}

	var bestOrder *api.SellOrderEntry = nil
	bestPrice := maxPrice
	for _, order := range *sellOrders {
		if order.Price <= bestPrice {
			bestOrder = &order
			bestPrice = order.Price
		}
	}

	// Price Check is done implicitly
	if bestOrder == nil {
		log(fmt.Sprintf("no sell orders for %s below %d gp", code, maxPrice))
		return nil, err
	}

	log(fmt.Sprintf("hitting sell order %s, buying %d %s at %d gp", bestOrder.Id, quantity, code, bestOrder.Price))
	res, err := actions.HitSellOrder(character, bestOrder.Id, quantity)
	if err != nil {
		log(fmt.Sprintf("failed to buy %s", code))
		return nil, err
	}

	utils.DebugLog(utils.PrettyPrint(res.Order))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
