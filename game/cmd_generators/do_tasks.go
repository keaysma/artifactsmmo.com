package generators

import (
	"fmt"
	"math"
	"strings"
	"time"

	"artifactsmmo.com/m/api"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/utils"
)

var simulationCount = 5
var totalFightCooldownThreshold = 7000 // 1800.0

func Tasks(task_type string) Generator {
	var retries = 0
	log := utils.LogPre(fmt.Sprintf("[tasks]<%s>", task_type))

	var initialized = false

	// Items Generator state
	var items_sub_generator *Generator = nil

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

		char := state.GlobalCharacter.Ref()
		characterName, characterHaste, current_task, current_task_type, task_progress, task_total := char.Name, char.Haste, char.Task, char.Task_type, char.Task_progress, char.Task_total
		x, y := char.X, char.Y
		hp, max_hp := char.Hp, char.Max_hp
		state.GlobalCharacter.Unlock()

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

		if task_progress >= task_total {
			items_sub_generator = nil

			// Put away any items we can
			// make sure we have enough space
			// for tasks_coins
			// next_command := DepositCheck(map[string]int{})
			// if next_command != "" {
			// 	return next_command
			// }

			itemCount := 0
			char := state.GlobalCharacter.Ref()
			for _, slot := range char.Inventory {
				itemCount += slot.Quantity
			}
			state.GlobalCharacter.Unlock()
			log(fmt.Sprintf("Has %d items", itemCount))

			if itemCount > 0 {
				return "deposit-everything"
			}

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
				fightResult, err := game.RunSimulations(characterName, current_task, simulationCount)
				if err != nil {
					log(fmt.Sprintf("Failed to get fight simulation results, abort: %s", err))
					return "clear-gen"
				}

				wins := 0
				for _, result := range *fightResult {
					if result.FightDetails.Result == "win" {
						wins++
					}
				}

				if wins < simulationCount {
					log(fmt.Sprintf("Won %d/%d fights, insufficiently successful, abort", wins, simulationCount))
					return "cancel-task"
				}

				cooldownSum := 0
				endHpSum := 0
				for _, result := range *fightResult {
					cooldown := game.GetCooldown(result.FightDetails.Turns, characterHaste)
					endHpSum += result.Metadata.CharacterEndHp
					log(fmt.Sprintf("cooldown: %d", cooldown))

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

				if totalCooldown > totalFightCooldownThreshold {
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
				monstersMaps, err = api.GetAllMapsByContentType("monster", current_task)
				if err != nil {
					log(fmt.Sprintf("failed to get maps for monster %s: %s", current_task, err))
					return "clear-gen"
				}
			}

			startsWithMove := strings.HasPrefix(last, "move")
			if !startsWithMove && last != "rest" && last != "fight" {
				closest_map := steps.PickClosestMap(coords.Coord{X: x, Y: y}, monstersMaps)
				move := fmt.Sprintf("move %d %d", closest_map.X, closest_map.Y)
				return move
			}

			if !steps.FightHpSafetyCheck(hp, max_hp) {
				return "rest"
			}

			return "fight"
		}

		if current_task_type == "items" {
			char := state.GlobalCharacter.Ref()
			task_item_count := utils.CountInventory(&char.Inventory, char.Task)
			state.GlobalCharacter.Unlock()

			// Turn in items if
			// - We're done with the task
			// Otherwise, "Make" runs its own Deposit and Withdraw checks
			if task_item_count >= task_total-task_progress {
				return "trade-task all"
			}

			// now we effectively need to sub-task the entire make or flip gen make
			if items_sub_generator == nil {
				log(fmt.Sprintf("building item generator for %s", current_task))
				generator := Make(current_task, true)
				items_sub_generator = &generator
			}

			return (*items_sub_generator)(last, success)
		}

		return "clear-gen" // drop-out
	}
}
