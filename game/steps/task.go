package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions/tasks"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func NewTask(kernel *game.Kernel, task_type string) (*types.Character, error) {
	log := kernel.LogPre(fmt.Sprintf("[%s]<task/new>: ", kernel.CharacterName))

	startChar := kernel.CharacterState.Ref()
	startTask, x, y := startChar.Task, startChar.X, startChar.Y
	kernel.CharacterState.Unlock()

	if startTask != "" {
		log(fmt.Sprintf("already has a task: %s", startTask))
		return nil, fmt.Errorf("already has a task: %s", startTask)
	}

	log("getting new task")
	maps, err := api.GetAllMapsByContentType("tasks_master", task_type)
	if err != nil {
		log(fmt.Sprintf("failed to get map info: %s", err))
		return nil, err
	}

	closest_map := PickClosestMap(coords.Coord{X: x, Y: y}, maps)
	_, err = Move(kernel, coords.Coord{X: closest_map.X, Y: closest_map.Y, Name: ""})
	if err != nil {
		log(fmt.Sprintf("failed to move to task master: %s", err))
	}

	res, err := tasks.NewTask(kernel.CharacterName)
	if err != nil {
		log(fmt.Sprintf("failed to get new task: %s", err))
		return nil, err
	}

	kernel.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Task)))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func TradeTaskItem(kernel *game.Kernel, quantitySelect BankDepositQuantityCb) (*types.Character, error) {
	log := kernel.LogPre(fmt.Sprintf("[%s]<task/trade>: ", kernel.CharacterName))

	startChar := kernel.CharacterState.Ref()
	startTask, startTaskType := startChar.Task, startChar.Task_type
	kernel.CharacterState.Unlock()

	if startTask == "" {
		log("does not have a task to complete!")
		return nil, fmt.Errorf("does not have a task to complete!")
	}

	if startTaskType != "items" {
		log(fmt.Sprintf("is not doing an items task, doing %s!", startTaskType))
		return nil, fmt.Errorf("is not doing an items task, doing %s!", startTaskType)
	}

	char := kernel.CharacterState.Ref()
	startTask, startTaskTotal, startTaskProgress := char.Task, char.Task_total, char.Task_progress
	startX, startY := char.X, char.Y
	inventorySlot := utils.FindInventorySlot(char, startTask)
	kernel.CharacterState.Unlock()

	current_count, quantity := 0, 0
	if inventorySlot != nil {
		current_count = inventorySlot.Quantity
		quantity = quantitySelect(*inventorySlot)
	} else {
		log(fmt.Sprintf("has no %s", startTask))
		return nil, fmt.Errorf("has no %s", startTask)
	}

	trade_quantity := min(quantity, current_count, startTaskTotal-startTaskProgress)

	maps, err := api.GetAllMapsByContentType("tasks_master", startTaskType)
	if err != nil {
		log(fmt.Sprintf("failed to get map info: %s", err))
		return nil, err
	}

	closestTile := PickClosestMap(coords.Coord{X: startX, Y: startY}, maps)
	_, err = Move(kernel, closestTile.IntoCoord())
	if err != nil {
		log(fmt.Sprintf("failed to move to task master: %s", err))
	}

	log(fmt.Sprintf("trades %d %s", trade_quantity, startTask))
	res, err := tasks.TradeTaskItem(kernel.CharacterName, startTask, trade_quantity)
	if err != nil {
		log(fmt.Sprintf("failed to trade task item: %s", err))
		return nil, err
	}

	kernel.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Trade)))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func CompleteTask(kernel *game.Kernel) (*types.Character, error) {
	log := kernel.LogPre(fmt.Sprintf("[%s]<task/complete>: ", kernel.CharacterName))

	char := kernel.CharacterState.Ref()
	startTask, startTaskType := char.Task, char.Task_type
	startX, startY := char.X, char.Y
	kernel.CharacterState.Unlock()

	if startTask == "" {
		log("does not have a task to complete!")
		return nil, fmt.Errorf("does not have a task to complete!")
	}

	log("completing task")
	maps, err := api.GetAllMapsByContentType("tasks_master", startTaskType)
	if err != nil {
		log(fmt.Sprintf("failed to get map info: %s", err))
		return nil, err
	}

	closestTile := PickClosestMap(coords.Coord{X: startX, Y: startY}, maps)
	_, err = Move(kernel, closestTile.IntoCoord())
	if err != nil {
		log(fmt.Sprintf("failed to move to task master: %s", err))
	}

	res, err := tasks.CompleteTask(kernel.CharacterName)
	if err != nil {
		log(fmt.Sprintf("failed to complete task: %s", err))
		return nil, err
	}

	kernel.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Reward)))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func CancelTask(kernel *game.Kernel) (*types.Character, error) {
	log := kernel.LogPre(fmt.Sprintf("[%s]<task/cancel>: ", kernel.CharacterName))

	char := kernel.CharacterState.Ref()
	startTask, startTaskType := char.Task, char.Task_type
	startX, startY := char.X, char.Y
	countTasksCoins := utils.CountInventory(&char.Inventory, "tasks_coin")
	kernel.CharacterState.Unlock()

	if startTask == "" {
		log("does not have a task to cancel!")
		return nil, fmt.Errorf("does not have a task to cancel!")
	}

	if countTasksCoins < 1 {
		bank_items, err := GetAllBankItems()
		if err != nil {
			log(fmt.Sprintf("failed to get bank items: %s", err))
			return nil, err
		}

		bank_tasks_coins_count := utils.CountBank(bank_items, "tasks_coin")
		if bank_tasks_coins_count < 1 {
			log(fmt.Sprintf("does not have enough tasks coins: %d", countTasksCoins))
			return nil, fmt.Errorf("not enough tasks coins to cancel task")
		}

		log("withdrawing tasks coins from bank")
		_, err = WithdrawBySelect(
			kernel,
			func(item types.InventoryItem) bool {
				return item.Code == "tasks_coin"
			},
			func(item types.InventoryItem) int {
				return 1
			},
		)

		if err != nil {
			log(fmt.Sprintf("failed to withdraw tasks coins: %s", err))
			return nil, err
		}
	}

	log("canceling task")
	maps, err := api.GetAllMapsByContentType("tasks_master", startTaskType)
	if err != nil {
		log(fmt.Sprintf("failed to get map info: %s", err))
		return nil, err
	}

	closestTile := PickClosestMap(coords.Coord{X: startX, Y: startY}, maps)
	_, err = Move(kernel, closestTile.IntoCoord())
	if err != nil {
		log(fmt.Sprintf("failed to move to task master: %s", err))
	}

	res, err := tasks.CancelTask(kernel.CharacterName)
	if err != nil {
		log(fmt.Sprintf("failed to cancel task: %s", err))
		return nil, err
	}

	kernel.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Reward)))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func ExchangeTaskCoins(kernel *game.Kernel) (*types.Character, error) {
	log := kernel.LogPre(fmt.Sprintf("[%s]<task/exchange>: ", kernel.CharacterName))

	char := kernel.CharacterState.Ref()
	startX, startY := char.X, char.Y
	countTasksCoins := utils.CountInventory(&char.Inventory, "tasks_coin")
	kernel.CharacterState.Unlock()

	if countTasksCoins < 6 {
		log(fmt.Sprintf("does not have enough tasks coins: %d", countTasksCoins))
		return nil, fmt.Errorf("does not have enough tasks coins: %d", countTasksCoins)
	}

	log("exchanging task coins")
	maps, err := api.GetAllMapsByContentType("tasks_master", "")
	if err != nil {
		log(fmt.Sprintf("failed to get map info: %s", err))
		return nil, err
	}

	closestTile := PickClosestMap(coords.Coord{X: startX, Y: startY}, maps)
	_, err = Move(kernel, closestTile.IntoCoord())
	if err != nil {
		log(fmt.Sprintf("failed to move to task master: %s", err))
	}

	res, err := tasks.ExchangeTaskCoins(kernel.CharacterName)
	if err != nil {
		log(fmt.Sprintf("failed to exchange task coins: %s", err))
		return nil, err
	}

	kernel.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Reward)))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
