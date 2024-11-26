package generators

import (
	"fmt"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

var INVENTORY_CLEAR_THRESHOLD = 1.0 // 0.9

func InventoryCheckLoop(countByResource map[string]int) string {
	log := utils.LogPre("<inv-check>: ")
	held_item_code_quantity_map := map[string]int{}

	char := state.GlobalCharacter.Ref()
	max_inventory_count := char.Inventory_max_items
	for _, slot := range char.Inventory {
		held_item_code_quantity_map[slot.Code] = slot.Quantity
	}
	state.GlobalCharacter.Unlock()

	current_inventory_count := 0
	for _, v := range held_item_code_quantity_map {
		current_inventory_count += v
	}

	if float64(current_inventory_count) < (float64(max_inventory_count) * INVENTORY_CLEAR_THRESHOLD) {
		// Inventory is not full
		return ""
	}

	task_item_count := held_item_code_quantity_map[char.Task]
	if task_item_count > 0 {
		return "trade-task all"
	}

	// Special case: Our inventory is full of auxiliary items
	// Time to put some stuff in the bank
	bank_inventory, err := api.GetBankItems()
	if err != nil {
		return "sleep 5" // hold-over, don't fail right now since alot of requests are being dropped by game server
	}

	for _, slot := range *bank_inventory {
		// special skip case: we can't deposit tasks_coin into the bank
		if slot.Code == "tasks_coin" {
			continue
		}

		_, needs := countByResource[slot.Code]
		quantity, has := held_item_code_quantity_map[slot.Code]
		if !needs && has && quantity > 0 {
			log(fmt.Sprintf("Depositing all %s", slot.Code))
			return fmt.Sprintf("deposit all %s", slot.Code)
		}
	}

	filled_bank_slots := 0
	for _, slot := range *bank_inventory {
		if slot.Code != "" {
			filled_bank_slots++
		}
	}

	bank_details, err := api.GetBankDetails()
	if err != nil {
		return "sleep 5" // hold-over, don't fail right now since alot of requests are being dropped by game server
	}

	allItemsHeldAreRequired := true
	if filled_bank_slots <= bank_details.Slots {
		// Special case: Our bank has plenty of space
		// Chuck the first item in the inventory into the bank
		for k, v := range held_item_code_quantity_map {
			if k == "" {
				continue
			}

			// do not deposit things we current need
			_, needs := countByResource[k]
			if needs {
				continue
			}

			log(fmt.Sprintf("doesnt need %s", k))
			allItemsHeldAreRequired = false

			if v > 0 {
				log(fmt.Sprintf("Depositing all %s", k))
				return fmt.Sprintf("deposit all %s", k)
			}
		}
	}

	if allItemsHeldAreRequired {
		return ""
	}

	// At this point
	// - The inventory > 90% full
	// - None of the held items are tradable for our task
	// - None of the held items are something we have a stack of in the bank
	// - We are holding some items we don't need for the current task (not sure if this is quite bug free, what if we're holding too many of one item?)
	// - The bank is > 90% full
	// I don't even know what I'd do manually at this point...
	// Human discretion is required, time to quit
	log("inventory full, no tradable items, no bank space")
	return "clear-gen"
}

func BuildResourceCountMap(component *steps.ItemComponentTree, resourceMap map[string]int, includeThisLevel bool) {
	// if component.Action == "gather" || component.Action == "fight" {
	if includeThisLevel {
		resourceMap[component.Code] = component.Quantity
	}

	for _, subcomponent := range component.Components {
		BuildResourceCountMap(&subcomponent, resourceMap, true)
	}
}

func BankCheck(data *steps.ItemComponentTree, countByResource map[string]int) string {
	log := utils.DebugLogPre("bank check: ")

	bank, err := api.GetBankItems()
	if err != nil {
		return ""
	}

	spaceRequiredPerCraft := 0
	for _, v := range countByResource {
		spaceRequiredPerCraft += v
	}

	currentInventoryCountByCode := map[string]int{}

	char := state.GlobalCharacter.Ref()
	defer state.GlobalCharacter.Unlock()

	// maxInventoryCount := int(float64(char.Inventory_max_items)*INVENTORY_CLEAR_THRESHOLD) - 1
	maxInventoryCount := char.Inventory_max_items
	inventoryCount := utils.CountAllInventory(char)
	freeSpace := maxInventoryCount - inventoryCount
	if freeSpace <= 0 {
		return ""
	}

	for _, slot := range char.Inventory {
		currentInventoryCountByCode[slot.Code] = slot.Quantity
	}

	for _, slot := range *bank {
		requiredAmount := countByResource[slot.Code]
		if requiredAmount <= 0 {
			continue
		}

		amountInBank := slot.Quantity
		log(fmt.Sprintf("amount in bank: %d", amountInBank))
		amountInInventory := currentInventoryCountByCode[slot.Code]
		log(fmt.Sprintf("amount in inv: %d", amountInInventory))

		maxCanMake := max(1, float64(freeSpace)/float64(spaceRequiredPerCraft))
		log(fmt.Sprintf("max can make: %f", maxCanMake))
		maxCanHold := int(maxCanMake) * requiredAmount
		log(fmt.Sprintf("max can hold: %d", maxCanHold))
		neededAmount := (maxCanHold - amountInInventory)
		log(fmt.Sprintf("needed: %d", neededAmount))

		amountToWithdraw := min(neededAmount, amountInBank, freeSpace)
		log(fmt.Sprintf("withdrawing: %d", amountToWithdraw))

		if amountToWithdraw > 0 {
			return fmt.Sprintf("withdraw %d %s", amountToWithdraw, slot.Code)
		}
	}

	return ""
}

func get_next_command_make(component *steps.ItemComponentTree, character *types.Character, skill_map *map[string]api.MapTile, last string, top bool) string {
	if !top && utils.CountInventory(&character.Inventory, component.Code) >= component.Quantity {
		return ""
	}

	if component.Action == "gather" || component.Action == "fight" {
		tile, ok := (*skill_map)[component.Code]
		if !ok {
			utils.Log(fmt.Sprintf("no map for resource %s", component.Code))
			return "clear-gen"
		}

		move := fmt.Sprintf("move %d %d", tile.X, tile.Y)

		utils.DebugLog(fmt.Sprintf("move: %s for %s %s", move, component.Action, component.Code))

		if last != move && last != component.Action && last != "rest" {
			return move
		}

		if component.Action == "fight" && !steps.FightHpSafetyCheck(character.Hp, character.Max_hp) {
			return "rest"
		}

		return component.Action
	}

	for _, subcomponent := range component.Components {
		next_command := get_next_command_make(&subcomponent, character, skill_map, last, false)
		if next_command != "" {
			return next_command
		}
	}

	return fmt.Sprintf("auto-craft %d %s", 1, component.Code) // component.Quantity
}

func Make(code string, doWithdrawFinishedFromBank bool) Generator {
	data, err := steps.GetItemComponentsTree(code)
	if err != nil {
		utils.Log(fmt.Sprintf("failed to get details on %s: %s", code, err))
		return Clear_gen
	}

	countByResource := map[string]int{}
	BuildResourceCountMap(data, countByResource, doWithdrawFinishedFromBank)

	var subtype_map = steps.ActionMap{}
	steps.BuildItemActionMapFromComponentTree(data, &subtype_map)

	resource_tile_map, err := steps.FindMapsForSubtypes(subtype_map)
	if err != nil {
		utils.Log(fmt.Sprintf("failed to get map info: %s", err))
		return Clear_gen
	}

	var retries = 0

	return func(last string, success bool) string {
		next_command := "clear-gen"

		if !success {
			// temporary - retry last command
			retries++
			if retries < 10 {
				return last
			}

			if retries < 15 {
				return "sleep 5"
			}

			return next_command
		}

		retries = 0

		next_command = InventoryCheckLoop(countByResource)
		if next_command != "" {
			return next_command
		}

		next_command = BankCheck(data, countByResource)
		if next_command != "" {
			return next_command
		}

		char := state.GlobalCharacter.Ref()
		next_command = get_next_command_make(data, char, resource_tile_map, last, true)
		state.GlobalCharacter.Unlock()

		// state.GlobalCharacter.With(func(value *types.Character) *types.Character {
		// 	next_command = get_next_command(data, value, resource_tile_map, last, true)
		// 	return value
		// })

		return next_command
	}
}
