package backend

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	coords "artifactsmmo.com/m/consts/places"
	generators "artifactsmmo.com/m/game/cmd_generators"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

var s = utils.GetSettings()

type InternalState struct {
	Generator_Paused     bool
	Current_Generator    generators.Generator
	Last_command         string
	Last_command_success bool
}

var internalState = InternalState{
	Generator_Paused:     false,
	Current_Generator:    nil,
	Last_command:         "",
	Last_command_success: false,
}

func ParseCommand(rawCommand string) bool {
	parts := strings.Split(rawCommand, " ")
	if len(parts) <= 0 {
		utils.Log("unparsable command")
		return true
	}

	command := parts[0]
	log := utils.LogPre(fmt.Sprintf("%s: ", command))
	switch command {
	case "ping":
		log("pong")
		return true
	case "sleep":
		if len(parts) != 2 {
			log("usage: sleep <seconds:number>")
			return false
		}

		raw_val := parts[1]
		sleep_time, err := strconv.ParseFloat(raw_val, 64)
		if err != nil {
			log(fmt.Sprintf("bad time value: %s", raw_val))
			return false
		}

		time.Sleep(time.Duration(sleep_time * float64(time.Second)))
		log(fmt.Sprintf("slept %f", sleep_time))
		return true
	case "move":
		if len(parts) != 3 {
			log("usage: move <x:number> <y:number>")
			return false
		}
		raw_x, raw_y := parts[1], parts[2]

		x, err := strconv.ParseInt(raw_x, 10, 64)
		if err != nil {
			log(fmt.Sprintf("can't parse x: %s", raw_x))
			return false
		}

		y, err := strconv.ParseInt(raw_y, 10, 64)
		if err != nil {
			log(fmt.Sprintf("can't parse y: %s", raw_y))
			return false
		}

		_, err = steps.Move(s.Character, coords.Coord{X: int(x), Y: int(y), Name: ""})
		if err != nil {
			log(fmt.Sprintf("failed to move to (%d, %d): %s", x, y, err))
			return false
		}

		return true
	case "use":
		if len(parts) != 3 {
			log("usage: use <quantity:number> <code:string>")
			return false
		}
		rawQuantity, code := parts[1], parts[2]

		quantity, err := strconv.ParseInt(rawQuantity, 10, 64)
		if err != nil {
			log(fmt.Sprintf("can't parse quantity: %s", rawQuantity))
			return false
		}

		_, err = steps.Use(s.Character, code, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to use %d %s: %s", quantity, code, err))
			return false
		}

		return true
	case "gather":
		_, err := steps.Gather(s.Character)
		if err != nil {
			log(fmt.Sprintf("failed to gather: %s", err))
			return false
		}

		return true
	case "fight":
		_, err := steps.Fight(s.Character)
		if err != nil {
			log(fmt.Sprintf("failed to fight: %s", err))
			return false
		}

		return true
	case "rest":
		_, err := steps.Rest(s.Character)
		if err != nil {
			log(fmt.Sprintf("failed to fight: %s", err))
			return false
		}

		return true
	case "equip":
		if len(parts) < 2 || len(parts) > 4 {
			log("usage: equip [<slot:string> ]<code:string>[ <quantity:number>]")
			return false
		}

		slot := ""
		code := parts[1]

		if len(parts) >= 3 {
			slot, code = parts[1], parts[2]
		}

		var err error
		var quantity int64 = 1
		if len(parts) == 4 {
			raw_quantity := parts[3]
			quantity, err = strconv.ParseInt(raw_quantity, 10, 64)
			if err != nil {
				log(fmt.Sprintf("can't parse quantity: %s", raw_quantity))
				return false
			}
		}

		err = steps.EquipItem(s.Character, code, slot, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to equip %d %s to %s: %s", quantity, code, slot, err))
			return false
		}

		return true
	case "unequip":
		if len(parts) < 2 || len(parts) > 3 {
			log("usage: unequip <slot:number>[ <quantity:number>]")
			return false
		}
		slot := parts[1]

		var err error
		var quantity int64 = 1
		if len(parts) == 3 {
			raw_quantity := parts[2]
			quantity, err = strconv.ParseInt(raw_quantity, 10, 64)
			if err != nil {
				log(fmt.Sprintf("can't parse quantity: %s", raw_quantity))
				return false
			}
		}

		err = steps.UnequipItem(s.Character, slot, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to unequip %d from %s: %s", quantity, slot, err))
			return false
		}

		return true
	case "buy":
		if len(parts) < 3 || len(parts) > 4 {
			log("usage: buy <quantity:number> <code:string>[ <max_price:number>]")
			return false
		}
		raw_quantity, code := parts[1], parts[2]
		quantity, err := strconv.ParseInt(raw_quantity, 10, 64)
		if err != nil {
			log(fmt.Sprintf("can't parse quantity: %s", raw_quantity))
			return false
		}

		var max_price int64 = -1
		if len(parts) == 4 {
			max_price_str := parts[3]
			max_price, err = strconv.ParseInt(max_price_str, 10, 64)
			if err != nil {
				log(fmt.Sprintf("can't parse max price: %s", max_price_str))
				return false
			}
		}

		_, err = steps.Buy(s.Character, code, int(quantity), int(max_price))
		if err != nil {
			log(fmt.Sprintf("failed to buy %d %s for price < %d: %s", quantity, code, max_price, err))
			return false
		}

		return true
	case "sell":
		if len(parts) < 3 || len(parts) > 4 {
			log("usage: sell <quantity:number or 'all'> <code:string>[ <min_price:number>]")
			return false
		}
		raw_quantity, code := parts[1], parts[2]

		var sellQuantity = 0
		if raw_quantity == "all" {
			state.GlobalCharacter.With(func(value *types.Character) *types.Character {
				sellQuantity = utils.CountInventory(&value.Inventory, code)
				return value
			})
		} else {
			quantity, err := strconv.ParseInt(raw_quantity, 10, 64)
			if err != nil {
				log(fmt.Sprintf("can't parse quantity: %s", raw_quantity))
				return false
			}
			sellQuantity = int(quantity)
		}

		var err error
		var min_price int64 = 0
		if len(parts) == 4 {
			min_price_str := parts[3]
			min_price, err = strconv.ParseInt(min_price_str, 10, 64)
			if err != nil {
				log(fmt.Sprintf("can't parse min price: %s", min_price_str))
				return false
			}
		}

		_, err = steps.Sell(s.Character, code, sellQuantity, int(min_price))
		if err != nil {
			log(fmt.Sprintf("failed to sell %s %s for price > %d: %s", raw_quantity, code, min_price, err))
			return false
		}

		return true
	case "o":
		// get orders for something particular
		// all-orders <code>

		if len(parts) != 2 {
			log("usage: o <code:string>")
			return false
		}

		code := parts[1]
		err := steps.ListSellOrders(code)
		if err != nil {
			log(fmt.Sprintf("failed to list orders for %s: %s", code, err))
			return false
		}

		return true
	case "myo":
		if len(parts) > 2 {
			log("usage: myo[ <code:string>]")
			return false
		}

		var logCode = "all"
		var code string = ""
		if len(parts) == 2 {
			code = parts[1]
			logCode = code
		}
		err := steps.ListMySellOrders(code)
		if err != nil {
			log(fmt.Sprintf("failed to list orders for %s: %s", logCode, err))
			return false
		}

		return true
	case "cancel-order":
		// cancel my orders on something
		// cancel-order <code> <id/all>
		if len(parts) != 2 {
			log("usage: cancel-order <id:string>")
			return false
		}

		id := parts[1]
		_, err := steps.CancelOrder(s.Character, id)
		if err != nil {
			log(fmt.Sprintf("failed to cancel order %s: %s", id, err))
			return false
		}

		return true
	case "hit-order":
		// buy a specific order
		// hit-order <id:string>[ <quantity:number>]
		if len(parts) < 2 || len(parts) > 3 {
			log("usage: hit-order <id:string>[ <quantity:number>]")
			return false
		}

		id := parts[1]
		var quantity int64 = -1
		if len(parts) == 3 {
			raw_quantity := parts[2]
			parsedQuantity, err := strconv.ParseInt(raw_quantity, 10, 64)
			if err != nil {
				log(fmt.Sprintf("can't parse quantity: %s", raw_quantity))
				return false
			}
			quantity = parsedQuantity
		}

		_, err := steps.HitOrder(s.Character, id, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to hit order %s: %s", id, err))
			return false
		}

		return true
	case "deposit":
		if len(parts) != 3 {
			log("usage: deposit <quantity:number or 'all'> <code:string>")
			return false
		}
		raw_quantity, code := parts[1], parts[2]
		quantity, _ := strconv.ParseInt(raw_quantity, 10, 64)

		_, err := steps.DepositBySelect(
			s.Character,
			func(item types.InventorySlot) bool {
				return item.Code == code
			},
			func(item types.InventorySlot) int {
				if raw_quantity == "all" {
					return item.Quantity
				}

				return int(quantity)
			},
		)
		if err != nil {
			log(fmt.Sprintf("failed to deposit %s %s: %s", raw_quantity, code, err))
			return false
		}

		return true
	case "withdraw":
		if len(parts) != 3 {
			log("usage: withdraw <quantity:number or 'all'> <code:string>")
			return false
		}
		raw_quantity, code := parts[1], parts[2]
		quantity, _ := strconv.ParseInt(raw_quantity, 10, 64)

		_, err := steps.WithdrawBySelect(
			s.Character,
			func(item types.InventoryItem) bool {
				return item.Code == code
			},
			func(item types.InventoryItem) int {
				if raw_quantity == "all" {
					return item.Quantity
				}

				return int(quantity)
			},
		)
		if err != nil {
			log(fmt.Sprintf("failed to withdraw %s %s: %s", raw_quantity, code, err))
			return false
		}

		return true
	case "deposit-gold":
		if len(parts) != 2 {
			log("usage: deposit-gold <quantity:number>")
			return false
		}

		raw_quantity := parts[1]
		quantity, err := strconv.ParseInt(raw_quantity, 10, 64)
		if err != nil {
			log(fmt.Sprintf("can't parse quantity: %s", raw_quantity))
			return false
		}

		_, err = steps.DepositGold(s.Character, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to deposit %d gold: %s", quantity, err))
			return false
		}

		return true
	case "withdraw-gold":
		if len(parts) != 2 {
			log("usage: withdraw-gold <quantity:number>")
			return false
		}

		raw_quantity := parts[1]
		quantity, err := strconv.ParseInt(raw_quantity, 10, 64)
		if err != nil {
			log(fmt.Sprintf("can't parse quantity: %s", raw_quantity))
			return false
		}

		_, err = steps.WithdrawGold(s.Character, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to withdraw %d gold: %s", quantity, err))
			return false
		}

		return true
	case "craft":
		if len(parts) < 2 || len(parts) > 3 {
			log("usage: craft[ <quantity:number>] <code:string>")
			return false
		}
		raw_quantity_or_code := parts[1]

		var quantity int64 = 1
		var code string = ""
		var err error

		if len(parts) == 2 {
			code = raw_quantity_or_code
		} else {
			code = parts[2]

			quantity, err = strconv.ParseInt(raw_quantity_or_code, 10, 64)
			if err != nil {
				log(fmt.Sprintf("can't parse quantity: %s", raw_quantity_or_code))
				return false
			}
		}

		_, err = steps.Craft(s.Character, code, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to craft %d %s: %s", quantity, code, err))
			return false
		}

		return true
	case "auto-craft":
		if len(parts) < 2 || len(parts) > 3 {
			log("usage: auto-craft[ <quantity:number>] <code:string>")
			return false
		}
		raw_quantity_or_code := parts[1]

		var quantity int64 = 1
		var code string = ""
		var err error

		if len(parts) == 2 {
			code = raw_quantity_or_code
		} else {
			code = parts[2]

			quantity, err = strconv.ParseInt(raw_quantity_or_code, 10, 64)
			if err != nil {
				log(fmt.Sprintf("can't parse quantity: %s", raw_quantity_or_code))
				return false
			}
		}

		_, err = steps.AutoCraft(s.Character, code, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to craft %s: %s", code, err))
			return false
		}

		return true
	case "new-task":
		if len(parts) != 2 {
			log("usage: new-task <type:'monsters'|'items'>")
			return false
		}
		task_type := parts[1]
		_, err := steps.NewTask(s.Character, task_type)
		if err != nil {
			log(fmt.Sprintf("failed to get new task: %s", err))
			return false
		}

		return true
	case "trade-task":
		if len(parts) != 2 {
			log("usage: trade-task <quantity:number or 'all'>")
			return false
		}
		raw_quantity := parts[1]
		quantity, _ := strconv.ParseInt(raw_quantity, 10, 64)

		_, err := steps.TradeTaskItem(
			s.Character,
			func(item types.InventorySlot) int {
				if raw_quantity == "all" {
					return item.Quantity
				}

				return int(quantity)
			},
		)
		if err != nil {
			log(fmt.Sprintf("failed to trade task item: %s", err))
			return false
		}

		return true
	case "complete-task":
		_, err := steps.CompleteTask(s.Character)
		if err != nil {
			log(fmt.Sprintf("failed to complete task: %s", err))
			return false
		}

		return true
	case "cancel-task":
		_, err := steps.CancelTask(s.Character)
		if err != nil {
			log(fmt.Sprintf("failed to cancel task: %s", err))
			return false
		}

		return true
	case "exchange-tasks-coins":
		_, err := steps.ExchangeTaskCoins(s.Character)
		if err != nil {
			log(fmt.Sprintf("failed to exchange task coins: %s", err))
			return false
		}

		return true
	case "gen":
		if len(parts) < 2 || len(parts) > 3 {
			log("usage: gen <name:string> <args:string>")
			return false
		}

		generator_name := parts[1]
		generator_arg := ""

		if len(parts) == 3 {
			generator_arg = parts[2]
		}

		success := true
		new_name := ""

		switch generator_name {
		case "level":
			if generator_arg == "" {
				log("missing generator argument")
				return false
			}
			internalState.Current_Generator = generators.Level(generator_arg)
			new_name = fmt.Sprintf("level <%s>", generator_arg)
		case "forever":
			if generator_arg == "" {
				log("missing generator argument")
				return false
			}
			internalState.Current_Generator = func(ctx string, success bool) string {
				if !success {
					return "clear-gen"
				}

				return generator_arg
			}
			new_name = fmt.Sprintf("forever <%s>", generator_arg)
		case "make":
			if generator_arg == "" {
				log("missing generator argument")
				return false
			}
			internalState.Current_Generator = generators.Make(generator_arg, false)
			new_name = fmt.Sprintf("make <%s>", generator_arg)
		case "flip":
			if generator_arg == "" {
				log("missing generator argument")
				return false
			}
			internalState.Current_Generator = generators.Flip(generator_arg)
			new_name = fmt.Sprintf("flip <%s>", generator_arg)
		case "tasks":
			if generator_arg == "" {
				log("missing generator argument")
				return false
			}
			internalState.Current_Generator = generators.Tasks(generator_arg)
			new_name = fmt.Sprintf("tasks <%s>", generator_arg)
		case "craft-sticky-sword":
			internalState.Current_Generator = generators.Craft_sticky_sword
			new_name = "craft-sticky-sword"
		case "fight-blue-slimes":
			internalState.Current_Generator = generators.Fight_blue_slimes
			new_name = "fight-blue-slimes"
		case "ashwood":
			internalState.Current_Generator = generators.Gather_ashwood
			new_name = "gather-ash-wood"
		case "dummy":
			internalState.Current_Generator = generators.Dummy
			new_name = "dummy"
		default:
			log(fmt.Sprintf("unknown generator: %s", generator_name))
			return false
		}

		if new_name != "" {
			log(fmt.Sprintf("generator set to %s", new_name))
			SharedState.Ref().Current_Generator_Name = new_name
			SharedState.Unlock()
		}

		return success
	case "pause-gen":
		internalState.Generator_Paused = true
		log("paused")
		return true
	case "resume-gen":
		internalState.Generator_Paused = false
		log("resuming")
		return true
	case "clear-gen":
		internalState.Current_Generator = nil
		SharedState.Ref().Current_Generator_Name = ""
		SharedState.Unlock()
		log("generator cleared")
		return true
	default:
		utils.Log(fmt.Sprintf("unknown command: %s", command))
		return false
	}
}

func Gameloop() {
	for {
		shared := SharedState.Ref()
		num_commands := len(shared.Commands)
		SharedState.Unlock()

		if num_commands == 0 {
			if internalState.Current_Generator != nil && !internalState.Generator_Paused {
				var c = internalState.Current_Generator(internalState.Last_command, internalState.Last_command_success)
				shared = SharedState.Ref()
				shared.Commands = append(shared.Commands, c)
				SharedState.Unlock()
			}
		} else {
			shared := SharedState.Ref()
			cmd, commands := shared.Commands[0], shared.Commands[1:]
			shared.Commands = commands
			SharedState.Unlock()

			is_success := ParseCommand(cmd)
			internalState.Last_command = cmd
			internalState.Last_command_success = is_success
		}

		// Nothing happened this loop,
		// Add a small sleep to prevent rapid looping
		time.Sleep(100 * time.Millisecond)
	}
}

func PriorityLoop(commands chan string) {
	for {
		// This loop is for high priority tasks
		// that need to be done immediately
		// stuff immune to cooldowns

		select {
		case cmd := <-commands:
			ParseCommand(cmd)
		default:
			time.Sleep(100 * time.Millisecond) // 100ms (0.1s)
		}
	}
}
