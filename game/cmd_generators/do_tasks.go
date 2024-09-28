package generators

import (
	"fmt"

	"artifactsmmo.com/m/api"
	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/utils"
)

func Tasks(task_type string) Generator {
	log := utils.LogPre(fmt.Sprintf("[tasks]<%s>", task_type))
	return func(last string, success bool) string {
		if !success {
			return "clear-gen"
		}

		char := state.GlobalCharacter.Ref()
		current_task, current_task_type, task_progress, task_total := char.Task, char.Task_type, char.Task_progress, char.Task_total
		x, y := char.X, char.Y
		state.GlobalCharacter.Unlock()

		if current_task == "" {
			// Allow breaking task loop
			if task_type == "finish" {
				return "clear-gen"
			}

			return fmt.Sprintf("new-task %s", task_type)
		}

		if task_progress >= task_total {
			return "complete-task"
		}

		// Regardless of what task_type is specified
		// finish the task that's of current_task_type
		if current_task_type == "monsters" {
			maps, err := api.GetAllMapsByContentType("monster", current_task)
			if err != nil {
				log(fmt.Sprintf("failed to get maps for monster %s: %s", current_task, err))
				return "clear-gen"
			}

			closest_map := steps.PickClosestMap(coords.Coord{X: x, Y: y}, maps)

			move := fmt.Sprintf("move %d %d", closest_map.X, closest_map.Y)
			if last != move && last != "fight" {
				return move
			}

			return "fight"
		}

		if current_task_type == "items" {
			// wip
		}

		return "clear-gen" // drop-out
	}
}
