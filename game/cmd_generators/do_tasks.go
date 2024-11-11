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
	var retries = 0
	var items_sub_generator *Generator = nil
	log := utils.LogPre(fmt.Sprintf("[tasks]<%s>", task_type))

	return func(last string, success bool) string {
		if !success {
			// temporary - retry last command
			retries++
			if retries < 10 {
				return last
			}

			return "clear-gen"
		}

		retries = 0

		char := state.GlobalCharacter.Ref()
		current_task, current_task_type, task_progress, task_total := char.Task, char.Task_type, char.Task_progress, char.Task_total
		x, y := char.X, char.Y
		hp, max_hp := char.Hp, char.Max_hp
		state.GlobalCharacter.Unlock()

		if current_task == "" {
			// Allow breaking task loop
			if task_type == "finish" {
				return "clear-gen"
			}

			return fmt.Sprintf("new-task %s", task_type)
		}

		if task_progress >= task_total {
			items_sub_generator = nil
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
			if last != move && last != "rest" && last != "fight" {
				return move
			}

			if max_hp-hp > 10 {
				return "rest"
			}

			return "fight"
		}

		if current_task_type == "items" {
			char := state.GlobalCharacter.Ref()
			task_item_count := utils.CountInventory(&char.Inventory, char.Task)
			max_inventory_count := char.Inventory_max_items
			current_inventory_count := utils.CountAllInventory(char)
			state.GlobalCharacter.Unlock()

			// Turn in items if
			// - We're done with the task
			// - The character inventory is getting close to full (>90% as an arbitrary "too full" point)
			// TODO: Just turning in task items doesn't guarantee the inventory won't fill up, need some kind-of inventory management handler still
			if task_item_count >= task_total-task_progress {
				return "trade-task all"
			}

			if float64(current_inventory_count) > (float64(max_inventory_count) * float64(0.9)) {
				if task_item_count > 0 {
					return "trade-task all"
				} else {
					// Special case: Our inventory is full of auxiliary items
					// Time to put some stuff in the bank
					bank_inventory, err := api.GetBankItems()
					if err != nil {
						return "sleep 5" // hold-over, don't fail right now since alot of requests are being dropped by game server
					}

					held_item_code_quantity_map := map[string]int{}
					char := state.GlobalCharacter.Ref()
					for _, slot := range char.Inventory {
						held_item_code_quantity_map[slot.Code] = slot.Quantity
					}
					state.GlobalCharacter.Unlock()

					for _, slot := range *bank_inventory {
						quantity, has := held_item_code_quantity_map[slot.Code]
						if has && quantity > 0 {
							return fmt.Sprintf("deposit all %s", slot.Code)
						}
					}

					// At this point
					// - The inventory > 90% full
					// - None of the held items are tradable for our task
					// - None of the held items are something we have a stack of in the bank
					// I don't even know what I'd do manually at this point...
					// Human discretion is required, time to quit
					log("inventory full, no tradable items, no bank space")
					return "clear-gen"
				}
			}

			// now we effectively need to sub-task the entire make or flip gen make
			if items_sub_generator == nil {
				log(fmt.Sprintf("building item generator for %s", current_task))
				generator := Make(current_task)
				items_sub_generator = &generator
			}

			return (*items_sub_generator)(last, success)
		}

		return "clear-gen" // drop-out
	}
}
