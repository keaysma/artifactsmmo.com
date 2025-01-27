package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions/tasks"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func NewTask(kernel *game.Kernel, task_type string) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<task/new>: ", kernel.CharacterName))

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

	utils.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Task)))
	kernel.CharacterState.Set(&res.Character)
	kernel.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func TradeTaskItem(kernel *game.Kernel, quantitySelect BankDepositQuantityCb) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<task/trade>: ", kernel.CharacterName))

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
	startTaskTotal, startTaskProgress := char.Task_total, char.Task_progress
	inventory_slot := utils.FindInventorySlot(char, startTask)
	kernel.CharacterState.Unlock()

	current_count, quantity := 0, 0
	if inventory_slot != nil {
		current_count = inventory_slot.Quantity
		quantity = quantitySelect(*inventory_slot)
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

	closest_map := PickClosestMap(coords.Coord{X: char_start.X, Y: char_start.Y}, maps)
	_, err = Move(character, coords.Coord{X: closest_map.X, Y: closest_map.Y, Name: ""})
	if err != nil {
		log(fmt.Sprintf("failed to move to task master: %s", err))
	}

	log(fmt.Sprintf("trades %d %s", trade_quantity, char_start.Task))
	res, err := tasks.TradeTaskItem(character, char_start.Task, trade_quantity)
	if err != nil {
		log(fmt.Sprintf("failed to trade task item: %s", err))
		return nil, err
	}

	utils.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Trade)))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func CompleteTask(kernel *game.Kernel) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<task/complete>: ", character))

	char_start, err := api.GetCharacterByName(character)
	if err != nil {
		log("failed to get character info")
		return nil, err
	}

	state.GlobalCharacter.Set(char_start)

	if char_start.Task == "" {
		log("does not have a task to complete!")
		return char_start, nil
	}

	log("completing task")
	maps, err := api.GetAllMapsByContentType("tasks_master", char_start.Task_type)
	if err != nil {
		log(fmt.Sprintf("failed to get map info: %s", err))
		return nil, err
	}

	closest_map := PickClosestMap(coords.Coord{X: char_start.X, Y: char_start.Y}, maps)
	_, err = Move(character, coords.Coord{X: closest_map.X, Y: closest_map.Y, Name: ""})
	if err != nil {
		log(fmt.Sprintf("failed to move to task master: %s", err))
	}

	res, err := tasks.CompleteTask(character)
	if err != nil {
		log(fmt.Sprintf("failed to complete task: %s", err))
		return nil, err
	}

	utils.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Reward)))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func CancelTask(kernel *game.Kernel) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<task/cancel>: ", character))

	char_start, err := api.GetCharacterByName(character)
	if err != nil {
		log("failed to get character info")
		return nil, err
	}

	state.GlobalCharacter.Set(char_start)

	if char_start.Task == "" {
		log("does not have a task to cancel!")
		return char_start, nil
	}

	tasks_coins_count := utils.CountInventory(&char_start.Inventory, "tasks_coin")
	if tasks_coins_count < 1 {
		bank_items, err := GetAllBankItems()
		if err != nil {
			log(fmt.Sprintf("failed to get bank items: %s", err))
			return nil, err
		}

		bank_tasks_coins_count := utils.CountBank(bank_items, "tasks_coin")
		if bank_tasks_coins_count < 1 {
			log(fmt.Sprintf("does not have enough tasks coins: %d", tasks_coins_count))
			return nil, fmt.Errorf("not enough tasks coins to cancel task")
		}

		log("withdrawing tasks coins from bank")
		_, err = WithdrawBySelect(
			character,
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
	maps, err := api.GetAllMapsByContentType("tasks_master", char_start.Task_type)
	if err != nil {
		log(fmt.Sprintf("failed to get map info: %s", err))
		return nil, err
	}

	closest_map := PickClosestMap(coords.Coord{X: char_start.X, Y: char_start.Y}, maps)
	_, err = Move(character, coords.Coord{X: closest_map.X, Y: closest_map.Y, Name: ""})
	if err != nil {
		log(fmt.Sprintf("failed to move to task master: %s", err))
	}

	res, err := tasks.CancelTask(character)
	if err != nil {
		log(fmt.Sprintf("failed to cancel task: %s", err))
		return nil, err
	}

	utils.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Reward)))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func ExchangeTaskCoins(kernel *game.Kernel) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<task/exchange>: ", character))

	char_start, err := api.GetCharacterByName(character)
	if err != nil {
		log("failed to get character info")
		return nil, err
	}

	state.GlobalCharacter.Set(char_start)

	taskCoinCount := utils.CountInventory(&char_start.Inventory, "tasks_coin")
	if taskCoinCount < 6 {
		log(fmt.Sprintf("does not have enough tasks coins: %d", taskCoinCount))
		return char_start, nil
	}

	log("exchanging task coins")
	maps, err := api.GetAllMapsByContentType("tasks_master", "")
	if err != nil {
		log(fmt.Sprintf("failed to get map info: %s", err))
		return nil, err
	}

	closest_map := PickClosestMap(coords.Coord{X: char_start.X, Y: char_start.Y}, maps)
	_, err = Move(character, coords.Coord{X: closest_map.X, Y: closest_map.Y, Name: ""})
	if err != nil {
		log(fmt.Sprintf("failed to move to task master: %s", err))
	}

	res, err := tasks.ExchangeTaskCoins(character)
	if err != nil {
		log(fmt.Sprintf("failed to exchange task coins: %s", err))
		return nil, err
	}

	utils.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Reward)))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}
