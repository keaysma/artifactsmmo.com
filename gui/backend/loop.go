package backend

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	generators "artifactsmmo.com/m/game/cmd_generators"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

const NUM_GENERATORS = 10

func NewKernel(character types.Character) *game.Kernel {
	cooldown := state.CooldownData{}

	// If this fails let's just ignore it, not critical
	end, err := time.Parse(time.RFC3339, character.Cooldown_expiration)
	if err == nil {
		cooldown.Duration_seconds = character.Cooldown
		cooldown.End = &end
	}

	internalState := game.Kernel{
		CharacterName:        character.Name,
		GeneratorPaused:      make([]bool, 10),
		Generators:           make([]*game.Generator, NUM_GENERATORS),
		Last_command:         "",
		Last_command_success: false,
		CurrentGeneratorName: utils.SyncData[string]{
			Value: "",
		},
		Commands: utils.SyncData[[]string]{
			Value: make([]string, 0),
		},
		PriorityCommands: make(chan string, 10),
		LogsChannel:      make(chan string, 100),

		// States
		CharacterState: utils.SyncData[types.Character]{
			Value: character,
		},
		CooldownState: utils.SyncData[state.CooldownData]{
			Value: cooldown,
		},

		// UI Shared
		BankItemListShown:  false,
		BankItemListFilter: nil,
	}

	return &internalState
}

func ParseCommand(kernel *game.Kernel, rawCommand string) bool {
	parts := strings.Split(rawCommand, " ")
	if len(parts) <= 0 {
		kernel.Log("unparsable command")
		return true
	}

	command := parts[0]
	log := kernel.LogPre(fmt.Sprintf("%s: ", command))
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

		_, err = steps.Move(kernel, coords.Coord{X: int(x), Y: int(y), Name: ""})
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

		err = steps.Use(kernel, code, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to use %d %s: %s", quantity, code, err))
			return false
		}

		return true
	case "gather":
		err := steps.Gather(kernel)
		if err != nil {
			log(fmt.Sprintf("failed to gather: %s", err))
			return false
		}

		return true
	case "fight":
		err := steps.Fight(kernel)
		if err != nil {
			log(fmt.Sprintf("failed to fight: %s", err))
			return false
		}

		return true
	case "fight-debug":
		err := steps.FightDebug(kernel)
		if err != nil {
			log(fmt.Sprintf("failed to fight: %s", err))
			return false
		}

		return true
	case "rest":
		_, err := steps.Rest(kernel)
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

		err = steps.EquipItem(kernel, code, slot, int(quantity))
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

		err = steps.UnequipItem(kernel, slot, int(quantity))
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

		err = steps.Buy(kernel, code, int(quantity), int(max_price))
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
			kernel.CharacterState.With(func(value *types.Character) *types.Character {
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

		err = steps.Sell(kernel, code, sellQuantity, int(min_price))
		if err != nil {
			log(fmt.Sprintf("failed to sell %s %s for price > %d: %s", raw_quantity, code, min_price, err))
			return false
		}

		return true
	case "list-bank":
		if len(parts) > 2 {
			log("usage: list-bank[ <code-partial:string>]")
			return false
		}

		if len(parts) == 2 {
			code_match := parts[1]
			kernel.BankItemListFilter = &code_match
		} else {
			kernel.BankItemListFilter = nil
		}

		_, err := steps.GetAllBankItems(true)
		if err != nil {
			log(fmt.Sprintf("failed to list bank items: %s", err))
			return false
		}

		kernel.BankItemListShown = true

		return true
	case "hide-bank":
		if len(parts) > 1 {
			log("usage: hide-bank")
			return false
		}

		kernel.BankItemListFilter = nil
		kernel.BankItemListShown = false

		return true
	case "o":
		// get orders for something particular
		// all-orders <code>

		if len(parts) != 2 {
			log("usage: o <code:string>")
			return false
		}

		code := parts[1]
		err := steps.ListSellOrders(kernel, code)
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
		err := steps.ListMySellOrders(kernel, code)
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
		err := steps.CancelOrder(kernel, id)
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

		err := steps.HitOrder(kernel, id, int(quantity))
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
			kernel,
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
	case "deposit-everything":
		if len(parts) != 1 {
			log("usage: deposit-everything")
			return false
		}

		_, err := steps.DepositEverything(kernel)
		if err != nil {
			log(fmt.Sprintf("failed to deposit everything: %s", err))
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
			kernel,
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

		_, err = steps.DepositGold(kernel, int(quantity))
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

		_, err = steps.WithdrawGold(kernel, int(quantity))
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

		_, err = steps.Craft(kernel, code, int(quantity))
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

		_, err = steps.AutoCraft(kernel, code, int(quantity))
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
		_, err := steps.NewTask(kernel, task_type)
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
			kernel,
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
		_, err := steps.CompleteTask(kernel)
		if err != nil {
			log(fmt.Sprintf("failed to complete task: %s", err))
			return false
		}

		return true
	case "cancel-task":
		_, err := steps.CancelTask(kernel)
		if err != nil {
			log(fmt.Sprintf("failed to cancel task: %s", err))
			return false
		}

		return true
	case "exchange-tasks-coins":
		_, err := steps.ExchangeTaskCoins(kernel)
		if err != nil {
			log(fmt.Sprintf("failed to exchange task coins: %s", err))
			return false
		}

		return true
	case "gen":
		if len(parts) < 2 || len(parts) > 4 {
			log("usage: gen [<number> ]<name:string>[ <args:string>]")
			return false
		}

		generator_name := ""
		generator_arg := ""
		var generator_number int64 = 0 // TODO

		generator_name_or_number := parts[1]
		maybe_generator_number, number_parse_err := strconv.ParseInt(generator_name_or_number, 10, 64)
		if number_parse_err != nil {
			generator_name = generator_name_or_number
		} else {
			generator_number = maybe_generator_number
			generator_name = parts[2]
		}

		if number_parse_err != nil {
			if len(parts) == 3 {
				generator_arg = parts[2]
			}
		} else {
			if len(parts) == 4 {
				generator_arg = parts[3]
			}
		}

		success := true
		new_name := ""
		var newGenerator game.Generator

		switch generator_name {
		case "level":
			if generator_arg == "" {
				log("missing generator argument")
				return false
			}
			newGenerator = generators.Level(kernel, generator_arg)
			new_name = fmt.Sprintf("level <%s>", generator_arg)
		case "forever":
			if generator_arg == "" {
				log("missing generator argument")
				return false
			}
			newGenerator = func(ctx string, success bool) string {
				if !success {
					return fmt.Sprintf("clear-gen %d", generator_number)
				}

				return generator_arg
			}
			new_name = fmt.Sprintf("forever <%s>", generator_arg)
		case "make":
			if generator_arg == "" {
				log("missing generator argument")
				return false
			}
			newGenerator = generators.Make(kernel, generator_arg, false)
			new_name = fmt.Sprintf("make <%s>", generator_arg)
		case "flip":
			if generator_arg == "" {
				log("missing generator argument")
				return false
			}
			newGenerator = generators.Flip(kernel, generator_arg)
			new_name = fmt.Sprintf("flip <%s>", generator_arg)
		case "tasks":
			if generator_arg == "" {
				log("missing generator argument")
				return false
			}
			newGenerator = generators.Tasks(kernel, generator_arg)
			new_name = fmt.Sprintf("tasks <%s>", generator_arg)
		case "fight-blue-slimes":
			newGenerator = generators.Fight_blue_slimes
			new_name = "fight-blue-slimes"
		case "ashwood":
			newGenerator = generators.Gather_ashwood
			new_name = "gather-ash-wood"
		case "dummy":
			newGenerator = generators.Dummy
			new_name = "dummy"
		default:
			log(fmt.Sprintf("unknown generator: %s", generator_name))
			return false
		}

		if new_name != "" {
			log(fmt.Sprintf("generator %d set to %s", generator_number, new_name))
			kernel.CurrentGeneratorName.Set(&new_name)
		}

		kernel.Generators[generator_number] = &newGenerator

		return success
	case "pause-gen":
		if len(parts) > 2 {
			log("usage: pause-gen[ <number>]")
			return false
		}

		if len(parts) == 2 {
			maybe_generator_number, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				log(fmt.Sprintf("failed to read number: %s", err))
				return false
			}

			kernel.GeneratorPaused[maybe_generator_number] = true
			log(fmt.Sprintf("paused %d", maybe_generator_number))
		} else {
			for i := range kernel.GeneratorPaused {
				kernel.GeneratorPaused[i] = true
			}
			log("paused all")
		}

		return true
	case "resume-gen":
		if len(parts) > 2 {
			log("usage: resume-gen[ <number>]")
			return false
		}

		if len(parts) == 2 {
			maybe_generator_number, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				log(fmt.Sprintf("failed to read number: %s", err))
				return false
			}

			kernel.GeneratorPaused[maybe_generator_number] = false
			log(fmt.Sprintf("resumed %d", maybe_generator_number))
		} else {
			for i := range kernel.GeneratorPaused {
				kernel.GeneratorPaused[i] = false
			}
			log("resumed all")
		}

		return true
	case "clear-gen":
		if len(parts) > 2 {
			log("usage: clear-gen[ <number>]")
			return false
		}

		if len(parts) == 2 {
			maybe_generator_number, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				log(fmt.Sprintf("failed to read number: %s", err))
				return false
			}

			kernel.Generators[maybe_generator_number] = nil
			log(fmt.Sprintf("cleared generator %d", maybe_generator_number))
		} else {
			kernel.Generators = make([]*game.Generator, NUM_GENERATORS)
			empty := ""
			kernel.CurrentGeneratorName.Set(&empty)
			log("cleared all generators")
		}

		return true
	case "simulate-fight":
		if len(parts) != 2 {
			log("usage: simulate-fight <monster_code:string>")
			return false
		}

		monster_code := parts[1]

		res, err := game.RunSimulations(kernel.CharacterName, monster_code, 1)
		if err != nil {
			log(fmt.Sprintf("failed to simulate fight: %s", err))
			return false
		}

		if len(*res) > 1 {
			wins, losses := 0, 0
			for _, fight := range *res {
				if fight.FightDetails.Result == "win" {
					wins++
				} else {
					losses++
				}
			}
			log(fmt.Sprintf("simulated fight: %d wins, %d losses", wins, losses))
		} else if len(*res) == 1 {
			for _, log := range (*res)[0].FightDetails.Logs {
				kernel.Log(log)
			}

			log(fmt.Sprintf("Cooldown: %d", (*res)[0].Metadata.Cooldown))
		} else {
			log("no results")
		}

		return true
	default:
		kernel.Log(fmt.Sprintf("unknown command: %s", command))
		return false
	}
}

func Gameloop(kernel *game.Kernel) {
	// zzz - respect already existing cooldown
	cd := kernel.CooldownState.ShallowCopy()
	remaining := time.Until(*cd.End)
	if remaining.Seconds() > 0 {
		time.Sleep(remaining)
	}

	for {
		commands := kernel.Commands.Ref()
		num_commands := len(*commands)
		kernel.Commands.Unlock()

		if num_commands == 0 {
			numGenerators := len(kernel.Generators)
			for i := range kernel.Generators {
				// reverse order
				idx := numGenerators - 1 - i
				paused := kernel.GeneratorPaused[idx]
				if paused == true {
					continue
				}

				gen := kernel.Generators[idx]
				if gen == nil {
					continue
				}

				var c = (*gen)(kernel.Last_command, kernel.Last_command_success)
				if c == "noop" {
					continue
				}

				kernel.Commands.With(func(value *[]string) *[]string {
					newValue := append(*value, c)
					return &newValue
				})
			}
		} else {
			commandsRef := kernel.Commands.Ref()
			commands := *commandsRef
			cmd, commands := commands[0], (commands)[1:]
			kernel.Commands.Value = commands
			kernel.Commands.Unlock()

			is_success := ParseCommand(kernel, cmd)
			kernel.Last_command = cmd
			kernel.Last_command_success = is_success
		}

		// Nothing happened this loop,
		// Add a small sleep to prevent rapid looping
		time.Sleep(100 * time.Millisecond)
	}
}

func PriorityLoop(kernel *game.Kernel) {
	for {
		// This loop is for high priority tasks
		// that need to be done immediately
		// stuff immune to cooldowns

		select {
		case cmd := <-kernel.PriorityCommands:
			ParseCommand(kernel, cmd)
		default:
			time.Sleep(100 * time.Millisecond) // 100ms (0.1s)
		}
	}
}
