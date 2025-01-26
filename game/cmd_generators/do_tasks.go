package generators

import (
	"fmt"
	"strings"
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/api/tasks"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/utils"
)

var quantityThreshold = 0.5
var simulationCount = 3
var turnCountThreshold = 10
var turnCountTotalThreshold = 0.5

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

			// If this happens its usually not network at this point
			// We have a task that we can't complete
			// We're stuck, time to quit
			// TODO: replace success bool with a more descriptive error
			return "cancel-task"
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
			next_command := DepositCheck(map[string]int{})
			if next_command != "" {
				return next_command
			}

			return "complete-task"
		}

		// Stuff that should only be done one time
		if !initialized {
			log("Initializing task")
			initialized = true

			if current_task_type == "monsters" {
				log("Monsters task, checking quantity")

				// Monsters tasks: Gauge if,
				// 1. Task amount is small enough
				// 2. Can fight monster reasonably fast
				taskDetails, err := tasks.GetTask(current_task)
				if err != nil {
					log(fmt.Sprintf("Failed to get task details, abort: %s", err))
					return "clear-gen"
				}

				quantityMidpoint := int(float64(taskDetails.Min_quantity+taskDetails.Max_quantity) * quantityThreshold)
				if task_total > quantityMidpoint {
					log(fmt.Sprintf("Task Total of %d is above limit of %d, cancel task", task_total, quantityMidpoint))
					return "cancel-task"
				}

				log("Amount is OK, check simulation result")
				fightResult, err := game.RunSimulations(characterName, current_task, simulationCount)
				if err != nil {
					log(fmt.Sprintf("Failed to get fight simulation results, abort: %s", err))
					return "clear-gen"
				}

				wins := 0
				turnCounts := []int{}
				for _, result := range *fightResult {
					if result.Result == "win" {
						wins++
					}

					turnCounts = append(turnCounts, result.Turns)
				}

				if wins < simulationCount {
					log(fmt.Sprintf("Won %d/%d fights, insufficiently successful, abort", wins, simulationCount))
					return "clear-gen"
				}

				turnsBelowThreshold := 0
				for _, turns := range turnCounts {
					if turns < turnCountThreshold {
						turnsBelowThreshold++
					}
					cooldown := game.GetCooldown(turns, characterHaste)
					log(fmt.Sprintf("cooldown: %d", cooldown))
				}

				turnCountOverUnder := float64(float64(turnsBelowThreshold) / float64(simulationCount))
				if turnCountOverUnder < turnCountTotalThreshold {
					log(fmt.Sprintf("Too many fights took too long, abort task %f < %f", turnCountOverUnder, turnCountTotalThreshold))
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
