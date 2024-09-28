package steps

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/actions/tasks"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

func NewTask(character string, task_type string) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<task/new>: ", character))

	char_start, err := api.GetCharacterByName(character)
	if err != nil {
		log(fmt.Sprintf("failed to get character info: %s", err))
		return nil, err
	}

	state.GlobalCharacter.Set(char_start)

	if char_start.Task != "" {
		log(fmt.Sprintf("already has a task: %s", char_start.Task))
		return char_start, nil
	}

	log("getting new task")
	maps, err := api.GetAllMapsByContentType("tasks_master", task_type)
	if err != nil {
		log(fmt.Sprintf("failed to get map info: %s", err))
		return nil, err
	}

	closest_map := PickClosestMap(coords.Coord{X: char_start.X, Y: char_start.Y}, maps)
	_, err = Move(character, coords.Coord{X: closest_map.X, Y: closest_map.Y, Name: ""})
	if err != nil {
		log(fmt.Sprintf("failed to move to task master: %s", err))
	}

	res, err := tasks.NewTask(character)
	if err != nil {
		log(fmt.Sprintf("failed to get new task: %s", err))
		return nil, err
	}

	utils.DebugLog(fmt.Sprintln(utils.PrettyPrint(res.Task)))
	state.GlobalCharacter.Set(&res.Character)
	api.WaitForDown(res.Cooldown)
	return &res.Character, nil
}

func CompleteTask(character string) (*types.Character, error) {
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

func ExchangeTaskCoins(character string) (*types.Character, error) {
	log := utils.LogPre(fmt.Sprintf("[%s]<task/exchange>: ", character))

	char_start, err := api.GetCharacterByName(character)
	if err != nil {
		log("failed to get character info")
		return nil, err
	}

	state.GlobalCharacter.Set(char_start)

	taskCoinCount := CountInventory(char_start, "tasks_coin")
	if taskCoinCount < 6 {
		log(fmt.Sprintf("does not have enough tasks coins: %d", taskCoinCount))
		return char_start, nil
	}

	log("exchanging task coins")
	// TODO: make me configurable monsters, items
	maps, err := api.GetAllMapsByContentType("tasks_master", "monsters")
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
