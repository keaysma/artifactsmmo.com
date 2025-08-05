package generators

import (
	"fmt"
	"math"
	"time"

	"artifactsmmo.com/m/api"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/fight_analysis"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/utils"
)

// var simulationCount = 10_000
var totalFightCooldownThreshold = 7000 // 1800.0

func Tasks(kernel *game.Kernel, task_type string) game.Generator {
	var retries = 0
	log := kernel.LogPre(fmt.Sprintf("[tasks]<%s>", task_type))

	var initialized = false

	// Items Generator state
	var items_sub_generator *game.Generator = nil

	// Monsters Generator state
	var monstersMaps *[]api.MapTile
	var err error

	return func(last string, success bool) string {
		if !success {
			// temporary - retry last command
			retries++

			// maybe its just a network issue
			if retries < 3 {
				return last
			}

			// ok... maybe the game server is down
			// give it a second...
			if retries < 10 {
				time.Sleep(5 * time.Second * time.Duration(retries))
				return last
			}

			return "clear-gen"

			// If this happens its usually not network at this point
			// We have a task that we can't complete
			// We're stuck, time to quit
			// TODO: replace success bool with a more descriptive error
			// return "cancel-task"
		}

		retries = 0

		characterName, characterHaste, current_task, current_task_type, task_progress, task_total := "", 0, "", "", 0, 0
		x, y := 0, 0
		hp, max_hp := 0, 0
		{
			char := kernel.CharacterState.Ref()
			characterName, characterHaste, current_task, current_task_type, task_progress, task_total = char.Name, char.Haste, char.Task, char.Task_type, char.Task_progress, char.Task_total
			x, y = char.X, char.Y
			hp, max_hp = char.Hp, char.Max_hp
			kernel.CharacterState.Unlock()
		}

		if current_task == "" {
			initialized = false
			items_sub_generator = nil
			monstersMaps = nil

			// Allow breaking task loop
			if task_type == "finish" {
				return "clear-gen"
			}

			return fmt.Sprintf("new-task %s", task_type)
		}

		// Put away any items we can
		// make sure we have enough space
		// for tasks_coin
		next_command := DepositCheck(kernel, map[string]int{})
		if next_command != "" {
			return next_command
		}

		if task_progress >= task_total {
			items_sub_generator = nil
			return "complete-task"
		}

		// Stuff that should only be done one time
		if !initialized {
			log("Initializing task")
			initialized = true

			if current_task_type == "monsters" {
				log(fmt.Sprintf("Monsters task %s, checking quantity", current_task))

				// Monsters tasks: Gauge if,
				// 1. Task amount is small enough
				// 2. Can fight monster reasonably fast

				log("Amount is OK, check simulation result")

				loadout, err := LoadOutForFight(kernel, current_task)
				if err != nil {
					log(fmt.Sprintf("Failed to get loadout for fight simulation: %s", err))
					return "clear-gen"
				}

				fightResult, err := fight_analysis.RunFightAnalysis(characterName, current_task, &loadout)
				if err != nil {
					log(fmt.Sprintf("Failed to get fight simulation results, abort: %s", err))
					return "clear-gen"
				}

				simulationCount := len((*fightResult).EndResults)
				wins := 0
				for _, result := range (*fightResult).EndResults {
					if result.CharacterWin {
						wins++
					}
				}

				if (float64(wins) / float64(simulationCount)) < 0.9 {
					log(fmt.Sprintf("Won %d/%d fights, insufficiently successful, abort", wins, simulationCount))
					return "cancel-task"
				}

				cooldownSum := 0
				endHpSum := 0
				for _, result := range (*fightResult).EndResults {
					cooldown := fight_analysis.GetCooldown(result.Turns, characterHaste)
					endHpSum += result.CharacterHp
					// log(fmt.Sprintf("cooldown: %d", cooldown))

					cooldownSum += cooldown
				}
				cooldownAvg := float64(cooldownSum) / float64(simulationCount)
				totalFightCooldown := int(cooldownAvg * float64(task_total))

				endHpAvg := int(math.Round(float64(endHpSum) / float64(simulationCount)))
				avgHpCost := max_hp - endHpAvg

				totalRestCooldown := 0
				if avgHpCost <= 0 {
					log("No HP cost, no rest needed!")
				} else {

					// im tired theres an easier way to do this with simple algebra but meh
					safetyHp := int(float64(max_hp) * steps.HP_SAFETY_PERCENT)
					fightingHp := max_hp
					fightsBeforeRest := 0
					for {
						if fightingHp < safetyHp {
							break
						}

						fightsBeforeRest++
						fightingHp -= avgHpCost
					}

					hpGainedDuringRestAvg := max_hp - fightingHp
					restCooldownAvg := int(math.Max(3, float64(hpGainedDuringRestAvg)/5.0))
					totalRests := float64(task_total) / float64(fightsBeforeRest)
					totalRestCooldown = int(totalRests * float64(restCooldownAvg))
				}

				totalCooldown := totalFightCooldown + totalRestCooldown

				log(fmt.Sprintf("Total fight cooldown (without rest estimates): %d", totalFightCooldown))
				log(fmt.Sprintf("Total fight cooldown (with rest estimates): %d", totalCooldown))

				tasksCoinCount := 0
				{
					bank := state.GlobalState.BankState.Ref()
					char := kernel.CharacterState.Ref()
					tasksCoinCount += utils.CountInventory(&char.Inventory, "task_coin")
					tasksCoinCount += utils.CountBank(bank, "task_coin")

					kernel.CharacterState.Unlock()
					state.GlobalState.BankState.Unlock()
				}

				if totalCooldown > totalFightCooldownThreshold && tasksCoinCount > 10 {
					log(fmt.Sprintf("Task will take too long, abort task %d > %d", totalCooldown, totalFightCooldownThreshold))
					return "cancel-task"
				}

			}

			log("Initialization OK")
		}

		// Regardless of what task_type is specified
		// finish the task that's of current_task_type
		if current_task_type == "monsters" {
			if monstersMaps == nil {
				monstersMaps, err = api.GetAllMaps("monster", current_task)
				if err != nil {
					log(fmt.Sprintf("failed to get maps for monster %s: %s", current_task, err))
					return "clear-gen"
				}
			}

			equipCommand, err := LoadOutCommand(kernel, current_task)
			if err != nil {
				log(fmt.Sprintf("failed to get equipment loadout for %s: %s", current_task, err))
				return "clear-gen"
			}

			if equipCommand != "" {
				return equipCommand
			}

			closest_map := steps.PickClosestMap(coords.Coord{X: x, Y: y}, monstersMaps)
			if x != closest_map.X || y != closest_map.Y {
				move := fmt.Sprintf("move %d %d", closest_map.X, closest_map.Y)
				return move
			}

			if !steps.FightHpSafetyCheck(hp, max_hp) {
				return "rest"
			}

			return "fight"
		}

		if current_task_type == "items" {
			char := kernel.CharacterState.Ref()
			task_item_count := utils.CountInventory(&char.Inventory, char.Task)
			kernel.CharacterState.Unlock()

			// Turn in items if
			// - We're done with the task
			// Otherwise, "Make" runs its own Deposit and Withdraw checks
			if task_item_count >= task_total-task_progress {
				return "trade-task all"
			}

			// now we effectively need to sub-task the entire make or flip gen make
			if items_sub_generator == nil {
				log(fmt.Sprintf("building item generator for %s", current_task))
				generator := Make(kernel, current_task, -1, true)
				items_sub_generator = &generator
			}

			return (*items_sub_generator)(last, success)
		}

		return "clear-gen" // drop-out
	}
}
