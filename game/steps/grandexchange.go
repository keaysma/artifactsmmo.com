package steps

import (
	"fmt"
	"strconv"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func ListSellOrders(kernel *game.Kernel, code string) error {
	logHead := kernel.LogPre("<ge/list-sell-orders> (o): ")
	log := kernel.LogPre("")

	orders, err := api.GetSellOrders(api.GetSellOrdersParams{
		Code: code,
	})
	if err != nil {
		logHead(fmt.Sprintf("failed to get sell orders for %s: %s", code, err))
		return err
	}

	ordersCache := []string{}
	for _, order := range *orders {
		log(fmt.Sprintf("%s: %d * %d gp = %d gp", order.Id, order.Quantity, order.Price, order.Quantity*order.Price))
		ordersCache = append(ordersCache, order.Id)
	}

	state.OrderIdsReference.Set(&ordersCache)

	history, err := api.GetSellOrderHistory(code, api.GetSellOrderHistoryParams{})
	if err != nil {
		logHead(fmt.Sprintf("failed to get sell order history for %s: %s", code, err))
		return err
	}

	for _, order := range (*history)[:min(len(*history), 10)] {
		log(fmt.Sprintf("%s: %d * %d gp = %d gp", order.Order_id, order.Quantity, order.Price, order.Quantity*order.Price))
	}

	return nil
}

func ListMySellOrders(kernel *game.Kernel, code string) error {
	logHead := kernel.LogPre("<ge/list-my-sell-orders> (o): ")
	log := kernel.LogPre("")

	orders, err := api.GetMySellOrders(
		api.GetMySellOrdersParams{
			Code: code,
		},
	)
	if err != nil {
		logHead(fmt.Sprintf("failed to get my sell orders: %s", err))
		return err
	}

	ordersCache := []string{}
	for _, order := range *orders {
		if code == "" {
			log(fmt.Sprintf("%s, %s: %d * %d gp = %d gp", order.Code, order.Id, order.Quantity, order.Price, order.Quantity*order.Price))
		} else {
			log(fmt.Sprintf("%s: %d * %d gp = %d gp", order.Id, order.Quantity, order.Price, order.Quantity*order.Price))
		}

		ordersCache = append(ordersCache, order.Id)
	}

	state.OrderIdsReference.Set(&ordersCache)

	return nil
}

func CancelOrder(kernel *game.Kernel, idMaybe string) error {
	log := kernel.LogPre("<ge/cancel-order>")

	// convert idMaybe to int
	id := idMaybe
	refNum, err := strconv.ParseInt(idMaybe, 10, 64)
	if err == nil {
		ordersCache := state.OrderIdsReference.Ref()
		if refNum < 0 || refNum >= int64(len(*ordersCache)) {
			log(fmt.Sprintf("invalid order reference %s", idMaybe))
			return err
		}
		id = (*ordersCache)[refNum]
		state.OrderIdsReference.Unlock()
	}

	log(fmt.Sprintf("getting ready to cancel order %s", id))

	_, err = Move(kernel, coords.GrandExchange)
	if err != nil {
		return err
	}

	res, err := actions.CancelSellOrder(kernel.CharacterName, id)
	if err != nil {
		log(fmt.Sprintf("failed to cancel order %s: %s", id, err))
		return err
	}

	kernel.DebugLog(utils.PrettyPrint(res.Order))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return nil
}

func Sell(kernel *game.Kernel, code string, quantity int, minPrice int) error {
	log := kernel.LogPre(fmt.Sprintf("[%s]<ge/sell>", kernel.CharacterName))

	_, err := Move(kernel, coords.GrandExchange)
	if err != nil {
		return err
	}

	sellOrders, err := api.GetSellOrders(
		api.GetSellOrdersParams{
			Code: code,
		},
	)
	if err != nil {
		log(fmt.Sprintf("failed to get sell orders for %s: %s", code, err))
		return err
	}

	sellOrderHistory, err := api.GetSellOrderHistory(
		code,
		api.GetSellOrderHistoryParams{},
	)
	if err != nil {
		log(fmt.Sprintf("failed to get sell order history for %s: %s", code, err))
		return err
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
		// return nil, err
	}

	var price = minPrice
	if minPrice == 0 {
		log(fmt.Sprintf("no min price specified, using historical average: %d", averageSellPrice))
		price = averageSellPrice
	}

	log(fmt.Sprintf("creating sell order for %d %s at %d gp", quantity, code, price))
	res, err := actions.CreateSellOrder(kernel.CharacterName, code, quantity, price)
	if err != nil {
		log(fmt.Sprintf("failed to sell %s", code))
		return err
	}

	kernel.DebugLog(utils.PrettyPrint(res.Order))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return nil
}

func Buy(kernel *game.Kernel, code string, quantity int, maxPrice int) error {
	log := kernel.LogPre(fmt.Sprintf("[%s]<ge/buy>", kernel.CharacterName))

	_, err := Move(kernel, coords.GrandExchange)
	if err != nil {
		return err
	}

	sellOrders, err := api.GetSellOrders(
		api.GetSellOrdersParams{
			Code: code,
		},
	)
	if err != nil {
		log(fmt.Sprintf("failed to get sell orders for %s: %s", code, err))
		return err
	}

	if len(*sellOrders) == 0 {
		log(fmt.Sprintf("no sell orders for %s", code))
		return err
	}

	var bestOrder *types.SellOrderEntry = nil
	bestPrice := maxPrice
	for _, order := range *sellOrders {
		if order.Price <= bestPrice || bestPrice < 0 {
			bestOrder = &order
			bestPrice = order.Price
		}
	}

	// Price Check is done implicitly
	if bestOrder == nil {
		log(fmt.Sprintf("no sell orders for %s below %d gp", code, maxPrice))
		return err
	}

	log(fmt.Sprintf("hitting sell order %s, buying %d %s at %d gp", bestOrder.Id, quantity, code, bestOrder.Price))
	res, err := actions.HitSellOrder(kernel.CharacterName, bestOrder.Id, quantity)
	if err != nil {
		log(fmt.Sprintf("failed to buy %s", code))
		return err
	}

	kernel.DebugLog(utils.PrettyPrint(res.Order))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return nil
}

func HitOrder(kernel *game.Kernel, idMaybe string, quantity int) error {
	log := kernel.LogPre("<ge/hit-order>")

	// convert idMaybe to int
	id := idMaybe
	refNum, err := strconv.ParseInt(idMaybe, 10, 64)
	if err == nil {
		ordersCache := state.OrderIdsReference.Ref()
		if refNum < 0 || refNum >= int64(len(*ordersCache)) {
			log(fmt.Sprintf("invalid order reference %s", idMaybe))
			return err
		}
		id = (*ordersCache)[refNum]
		state.OrderIdsReference.Unlock()
	}

	_, err = Move(kernel, coords.GrandExchange)
	if err != nil {
		return err
	}

	orderQuantity := quantity
	if orderQuantity < 0 {
		info, err := api.GetSellOrder(id)
		if err != nil {
			log(fmt.Sprintf("failed to get order %s: %s", id, err))
			return err
		}

		log(fmt.Sprintf("order %s has %d items", id, info.Quantity))

		orderQuantity = info.Quantity
	}

	res, err := actions.HitSellOrder(kernel.CharacterName, id, orderQuantity)
	if err != nil {
		log(fmt.Sprintf("failed to hit order  %s: %s", id, err))
		return err
	}

	kernel.DebugLog(utils.PrettyPrint(res.Order))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return nil
}
