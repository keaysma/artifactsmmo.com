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
var BANK_CLEAR_THRESHOLD = 1.0

func DepositCheck(needsCodeQuantity map[string]int) string {
	log := utils.LogPre("<deposit-check>: ")
	heldCodeQuantity := map[string]int{}

	char := state.GlobalCharacter.Ref()
	currentFilledSlots := 0
	totalSlots := len(char.Inventory)
	currentInventoryCount := 0
	maxInventoryCount := char.Inventory_max_items
	for _, slot := range char.Inventory {
		heldCodeQuantity[slot.Code] = slot.Quantity
		currentInventoryCount += slot.Quantity
		if slot.Code != "" {
			currentFilledSlots++
		}
	}
	state.GlobalCharacter.Unlock()

	if float64(currentInventoryCount) < (float64(maxInventoryCount)*INVENTORY_CLEAR_THRESHOLD) && currentFilledSlots < totalSlots {
		// Inventory is not full
		// Skip this check
		return ""
	}

	// Special case: Our inventory is full of auxiliary items
	// Time to put some stuff in the bank
	bankItems, err := steps.GetAllBankItems()
	if err != nil {
		log(fmt.Sprintf("Check failed, fetching bank items was unsuccessful: %s", err))
		return "sleep 5"
	}

	maxBankCount := len(*bankItems)
	currentBankCount := 0
	bankCodeQuantity := map[string]int{}
	for _, slot := range *bankItems {
		if slot.Code != "" {
			currentBankCount++
			bankCodeQuantity[slot.Code] = slot.Quantity
		}
	}
	bankIsFull := float64(currentBankCount) >= (float64(maxBankCount) * BANK_CLEAR_THRESHOLD)

	// First try to get rid of things we do not need
	for code := range heldCodeQuantity {
		_, needs := needsCodeQuantity[code]
		if !needs {
			if bankIsFull {
				_, bankHas := bankCodeQuantity[code]
				if !bankHas {
					// skip me, the bank is too full
					continue
				}
			}
			log(fmt.Sprintf("Junk item, depositing all %s", code))
			return fmt.Sprintf("deposit all %s", code)
		}
	}

	// If nothing was otherwise deposited, this is a good time to turn in task items
	task_item_count := heldCodeQuantity[char.Task]
	if task_item_count > 0 {
		return "trade-task all"
	}

	// Now try to reduce overstock on items we need, but are holding too much of
	// At this point, we are only holding things we need
	// so we can assume that we're able maximize craft amounts,
	// so anything more than required for the max amounts of crafts needs to go
	spaceRequiredPerCraft := 0
	for _, v := range needsCodeQuantity {
		spaceRequiredPerCraft += v
	}
	maxCanMake := max(1, float64(maxInventoryCount)/float64(spaceRequiredPerCraft))
	for code, quantity := range heldCodeQuantity {
		maxNeeded := int(maxCanMake * float64(needsCodeQuantity[code]))
		excessHeldQuantity := quantity - maxNeeded
		if excessHeldQuantity > 0 {
			log(fmt.Sprintf("Need %s, but holding %d in excess, depositing", code, quantity))
			return fmt.Sprintf("deposit %d %s", quantity, code)
		}
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

func WithdrawCheck(needsCodeQuantity map[string]int, targetItemCode string) string {
	// targetItemCode is the item we're trying to make
	// do not consider it when calculating how much space we need
	// to make that item

	log := utils.LogPre("<withdraw-check>: ")
	heldCodeQuantity := map[string]int{}

	spaceRequiredPerCraft := 0
	for code, v := range needsCodeQuantity {
		if code == targetItemCode {
			continue
		}

		spaceRequiredPerCraft += v
	}

	// Incase there are 0 requirements for the item
	spaceRequiredPerCraft = max(1, spaceRequiredPerCraft)

	char := state.GlobalCharacter.Ref()
	currentInventoryCount := 0
	maxInventoryCount := char.Inventory_max_items
	for _, slot := range char.Inventory {
		heldCodeQuantity[slot.Code] = slot.Quantity
		currentInventoryCount += slot.Quantity
	}
	state.GlobalCharacter.Unlock()

	freeSpace := maxInventoryCount - currentInventoryCount
	log(fmt.Sprintf("free space: %d", freeSpace))
	if freeSpace <= 0 {
		return ""
	}

	maxCanMake := max(1, float64(freeSpace)/float64(spaceRequiredPerCraft))
	log(fmt.Sprintf("max can make: %f", maxCanMake))

	bankItems, err := steps.GetAllBankItems()
	if err != nil {
		log(fmt.Sprintf("Check failed, fetching bank items was unsuccessful: %s", err))
		return "sleep 5"
	}

	for _, slot := range *bankItems {
		needsPerQuantity, needs := needsCodeQuantity[slot.Code]
		if !needs {
			continue
		}

		storedQuantity := slot.Quantity
		log(fmt.Sprintf("%s amount in bank: %d", slot.Code, storedQuantity))

		heldQuantity := heldCodeQuantity[slot.Code]
		log(fmt.Sprintf("%s amount in inv: %d", slot.Code, heldQuantity))

		targetQuantity := int(maxCanMake) * needsPerQuantity
		log(fmt.Sprintf("%s target: %d", slot.Code, targetQuantity))

		maxWithdrawQuantity := targetQuantity - heldQuantity
		log(fmt.Sprintf("%s wants to withdraw: %d", slot.Code, maxWithdrawQuantity))

		amountToWithdraw := min(maxWithdrawQuantity, storedQuantity, freeSpace)
		if amountToWithdraw > 0 {
			log(fmt.Sprintf("%s withdrawing: %d", slot.Code, amountToWithdraw))
			return fmt.Sprintf("withdraw %d %s", amountToWithdraw, slot.Code)
		}
	}

	return ""
}

func BuildResourceCountMap(component *steps.ItemComponentTree, resourceMap map[string]int, includeThisLevel bool) {
	if includeThisLevel {
		resourceMap[component.Code] = component.Quantity
	}

	for _, subcomponent := range component.Components {
		BuildResourceCountMap(&subcomponent, resourceMap, true)
	}
}

func NextMakeAction(component *steps.ItemComponentTree, character *types.Character, skill_map *map[string]api.MapTile, last string, top bool) string {
	log := utils.DebugLogPre("<next-make-action>: ")

	if !top && utils.CountInventory(&character.Inventory, component.Code) >= component.Quantity {
		return ""
	}

	if component.Action == "gather" || component.Action == "fight" || component.Action == "withdraw" {
		tile, ok := (*skill_map)[component.Code]
		if !ok {
			utils.Log(fmt.Sprintf("no map for resource %s", component.Code))
			return "cancel-task"
		}

		if tile.Content.Type == "event" {
			// want to skip events for now
			// they tend to take too much time
			return "cancel-task"

			/*
				utils.Log(fmt.Sprintf("find event tile for resource %s", component.Code))
				events, err := api.GetAllActiveEvents(1, 100)
				if err != nil {
					utils.Log(fmt.Sprintf("failed to get event info: %s", err))
					return "sleep 10"
				}

				if len(*events) == 0 {
					utils.Log(fmt.Sprintf("no event info found for %s", component.Code))
					return "sleep 10"
				}

				didFindActiveEvent := false
				for _, event := range *events {
					if event.Map.Content.Code == tile.Content.Code {
						didFindActiveEvent = true
						utils.Log(fmt.Sprintf("event: %s", event.Code))
						if character.X != event.Map.X || character.Y != event.Map.Y {
							return fmt.Sprintf("move %d %d", event.Map.X, event.Map.Y)
						}
					}
				}

				if !didFindActiveEvent {
					utils.Log(fmt.Sprintf("no active events for %s, tile %s - sleep", component.Code, tile.Content.Code))
					return "sleep 10"
				}
			*/
		} else if character.X != tile.X || character.Y != tile.Y {
			move := fmt.Sprintf("move %d %d", tile.X, tile.Y)
			utils.DebugLog(fmt.Sprintf("move: %s for %s %s", move, component.Action, component.Code))
			return move
		}

		switch component.Action {
		case "gather":
			return "gather"
		case "fight":
			if steps.FightHpSafetyCheck(character.Hp, character.Max_hp) {
				return "fight"
			} else {
				return "rest"
			}
		case "withdraw":
			withdraw := fmt.Sprintf("withdraw %d %s", component.Quantity, component.Code)
			return withdraw
		default:
			log(fmt.Sprintf("HOW DID WE GET HERE??? action is %s", component.Action))
			return "clear-gen"
		}
	}

	for _, subcomponent := range component.Components {
		next_command := NextMakeAction(&subcomponent, character, skill_map, last, false)
		if next_command != "" {
			return next_command
		}
	}

	return fmt.Sprintf("auto-craft %d %s", 1, component.Code) // component.Quantity
}

func Make(code string, needsFinishedItem bool) Generator {
	// needsFinishedItem:
	// ... sometimes we want to permit putting the finished product in the bank (we're skilling, need more space to make more items)
	// ... other times we need to hold on to that finished item (we're doing tasks, need to turn these finished items in to the task master)

	data, err := steps.GetItemComponentsTree(code)
	if err != nil {
		utils.Log(fmt.Sprintf("failed to get details on %s: %s", code, err))
		return Clear_gen
	}

	countByResource := map[string]int{}
	BuildResourceCountMap(data, countByResource, needsFinishedItem)

	var mapCodeAction = steps.ActionMap{}
	steps.BuildItemActionMapFromComponentTree(data, &mapCodeAction)

	resource_tile_map, err := steps.FindMapsForActions(mapCodeAction)
	if err != nil || resource_tile_map == nil || len(*resource_tile_map) == 0 {
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

			// If this happens its usually not network at this point
			// We have a task that we can't complete
			// We're stuck, time to quit
			// TODO: replace success bool with a more descriptive error
			return "cancel-task"
			//return next_command
		}

		retries = 0

		next_command = DepositCheck(countByResource)
		if next_command != "" {
			return next_command
		}

		next_command = WithdrawCheck(countByResource, code)
		if next_command != "" {
			return next_command
		}

		char := state.GlobalCharacter.Ref()
		next_command = NextMakeAction(data, char, resource_tile_map, last, true)
		state.GlobalCharacter.Unlock()

		// state.GlobalCharacter.With(func(value *types.Character) *types.Character {
		// 	next_command = get_next_command(data, value, resource_tile_map, last, true)
		// 	return value
		// })

		return next_command
	}
}
