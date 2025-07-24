package backend

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"artifactsmmo.com/m/api"
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
		CharacterName:   character.Name,
		GeneratorPaused: make([]bool, 10),
		Generators:      make([]*game.Generator, NUM_GENERATORS),
		GeneratorNames: utils.SyncData[[]string]{
			Value: make([]string, 10),
		},
		Last_command:         "",
		Last_command_success: false,
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
	log := kernel.LogPreF(fmt.Sprintf("%s: ", command))
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
		var quantity int64 = 0
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
	case "loadout":
		if len(parts) < 2 {
			log("usage: equip [<slot:string>:]<code:string>[:<quantity:number>] [...]")
			return false
		}

		type eqset struct {
			Slot     string
			Code     string
			Quantity int
		}
		equipSets := []eqset{}

		for _, pair := range parts[1:] {
			sec := strings.Split(pair, ":")

			eq := eqset{}

			if len(sec) == 1 {
				eq.Code = sec[0]
			} else if len(sec) == 2 {
				// This is either:
				// code:quantity
				// slot:code

				maybeQt, err := strconv.ParseInt(sec[1], 10, 64)
				if err == nil {
					// it is code:quantity
					eq.Code = sec[0]
					eq.Quantity = int(maybeQt)
				} else {
					// it is slot:code
					eq.Slot = sec[0]
					eq.Code = sec[1]
				}
			} else if len(sec) == 3 {
				maybeQt, err := strconv.ParseInt(sec[2], 10, 64)
				if err != nil {
					log("usage: equip [<slot:string>:]<code:string>[:<quantity:number>] [...]")
					return false
				}

				eq.Slot = sec[0]
				eq.Code = sec[1]
				eq.Quantity = int(maybeQt)
			} else {
				log("usage: equip [<slot:string>:]<code:string>[:<quantity:number>] [...]")
				return false
			}

			equipSets = append(equipSets, eq)
		}

		for _, eq := range equipSets {
			err := steps.EquipItem(kernel, eq.Code, eq.Slot, eq.Quantity)
			if err != nil {
				log(fmt.Sprintf("failed to equip %d %s to %s: %s", eq.Quantity, eq.Code, eq.Slot, err))
				return false
			}
		}

		return true
	case "unequip":
		if len(parts) < 2 || len(parts) > 3 {
			log("usage: unequip <slot:string>[ <quantity:number>]")
			return false
		}
		slot := parts[1]
		switch slot {
		case "u1":
			slot = "utility1"
		case "u2":
			slot = "utility2"
		case "a1":
			slot = "artifact1"
		case "a2":
			slot = "artifact2"
		case "a3":
			slot = "artifact3"
		case "r1":
			slot = "ring1"
		case "r2":
			slot = "ring2"
		}

		var err error
		var quantity int64 = 0
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
	case "npc-buy":
		if len(parts) != 3 {
			log("usage: npc-buy <quantity:number> <code:string>")
			return false
		}
		raw_quantity, code := parts[1], parts[2]
		quantity, err := strconv.ParseInt(raw_quantity, 10, 64)
		if err != nil {
			log(fmt.Sprintf("can't parse quantity: %s", raw_quantity))
			return false
		}

		err = steps.NPCBuy(kernel, code, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to buy %d %s: %s", quantity, code, err))
			return false
		}

		return true
	case "npc-sell":
		if len(parts) != 3 {
			log("usage: npc-sell <quantity:number> <code:string>")
			return false
		}
		raw_quantity, code := parts[1], parts[2]
		quantity, err := strconv.ParseInt(raw_quantity, 10, 64)
		if err != nil {
			log(fmt.Sprintf("can't parse quantity: %s", raw_quantity))
			return false
		}

		err = steps.NPCSell(kernel, code, int(quantity))
		if err != nil {
			log(fmt.Sprintf("failed to sell %d %s: %s", quantity, code, err))
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
		if len(parts) < 2 {
			log("usage: gen [<number> ]<name:string>[ <arg0:string>[ <arg1:string>]]")
			return false
		}

		generator_name := ""
		generator_args := []string{""}
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
			if len(parts) >= 3 {
				generator_args = parts[2:]
			}
		} else {
			if len(parts) >= 4 {
				generator_args = parts[3:]
			}
		}

		success := true
		new_name := ""
		var newGenerator game.Generator

		switch generator_name {
		case "level":
			if generator_args[0] == "" {
				log("missing generator argument")
				return false
			}

			var count int = -1
			if len(generator_args) > 1 {
				maybe_count, err := strconv.ParseInt(generator_args[1], 10, 64)
				if err == nil && maybe_count != 0 {
					count = int(maybe_count)
				}
			}

			newGenerator = generators.Level(kernel, generator_args[0], count)
			new_name = fmt.Sprintf("level <%s>", generator_args[0])
		case "forever":
			if generator_args[0] == "" {
				log("missing generator argument")
				return false
			}
			i := 0
			newGenerator = func(ctx string, success bool) string {
				if !success {
					return "clear-gen"
				}

				i++
				if i >= len(generator_args) {
					i = 0
				}

				return strings.Replace(generator_args[i], "~", " ", -1)
			}
			new_name = fmt.Sprintf("forever <%s>", generator_args[0])
		case "make":
			if generator_args[0] == "" {
				log("missing generator argument")
				return false
			}

			var count int = -1
			if len(generator_args) > 1 {
				maybe_count, err := strconv.ParseInt(generator_args[1], 10, 64)
				if err == nil && maybe_count != 0 {
					count = int(maybe_count)
				}
			}

			newGenerator = generators.Make(kernel, generator_args[0], count, false)
			new_name = fmt.Sprintf("make <%s>", generator_args[0])
		case "flip":
			if generator_args[0] == "" {
				log("missing generator argument")
				return false
			}
			newGenerator = generators.Flip(kernel, generator_args[0])
			new_name = fmt.Sprintf("flip <%s>", generator_args[0])
		case "tasks":
			if generator_args[0] == "" {
				log("missing generator argument")
				return false
			}
			newGenerator = generators.Tasks(kernel, generator_args[0])
			new_name = fmt.Sprintf("tasks <%s>", generator_args[0])
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
			kernel.GeneratorNames.With(func(value *[]string) *[]string {
				(*value)[generator_number] = new_name
				return value
			})
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
			kernel.GeneratorNames.With(func(value *[]string) *[]string {
				(*value)[maybe_generator_number] = ""
				return value
			})
			log(fmt.Sprintf("cleared generator %d", maybe_generator_number))
		} else {
			kernel.Generators = make([]*game.Generator, NUM_GENERATORS)
			kernel.GeneratorNames.With(func(value *[]string) *[]string {
				new := make([]string, 10)
				return &new
			})
			log("cleared all generators")
		}

		return true
	case "best":
		if len(parts) < 2 {
			log("usage: best <type:string>[ <level-range>][ <effect0:equation> [...]]")
			log("usage: best <type:string> against <monster:string>[ given <weapon:string>]")
			log("usage: best loadout against <monster:string>[ <algo:'v2'|'DAG'|'sim'>]")
			return false
		}

		search_type := parts[1]

		// ... kind of hacky
		if len(parts) >= 4 && parts[2] == "against" {
			// 'against' functionality
			target := parts[3]

			// assumed level range of 0 -> current level
			currentLevel := 0
			kernel.CharacterState.Read(func(value *types.Character) { currentLevel = value.Level })

			if search_type == "loadout" {
				algoName := "v2"
				if len(parts) >= 5 {
					algoName = parts[4]
				}
				algo := generators.LoadOutForFightV2
				switch algoName {
				case "v1":
					algo = generators.LoadOutForFightV1
				case "v2":
					algo = generators.LoadOutForFightV2
				case "v3", "dag":
					algo = generators.LoadOutForFightDAG
				case "v4", "sim":
					algo = generators.LoadOutForFightBruteForce
				case "v5", "prb":
					algo = generators.LoadOutForFightAnalysis
				default:
					log(fmt.Sprintf("unknown algo: %s", algoName))
					return false
				}

				log(fmt.Sprintf("find the best loadout, target=%s, max_level=%d, algo=%s", target, currentLevel, algoName))
				res, err := algo(kernel, target)
				if err != nil {
					log(fmt.Sprintf("Failed to get loadout: %s", err))
					return false
				}
				if len(res) > 0 {
					for slot, item := range res {
						char := kernel.CharacterState.Ref()
						curItem := utils.GetFieldFromStructByName(char, fmt.Sprintf("%s_slot", slot)).String()
						kernel.CharacterState.Unlock()

						if curItem == item.Code {
							continue
						}

						log(fmt.Sprintf("%s=%s", slot, item.Code))
					}
				} else {
					log("current loadout is best")
				}

				return true
			} else {
				selectedWeapon := ""
				if len(parts) == 6 && parts[4] == "given" {
					selectedWeapon = parts[5]
				} else {
					kernel.CharacterState.Read(func(value *types.Character) {
						selectedWeapon = value.Weapon_slot
					})
				}

				var dmgCtx *[]types.Effect = nil
				details, err := api.GetItemDetails(selectedWeapon)
				if err != nil {
					log(fmt.Sprintf("error getting item details for weapon %s: %s", selectedWeapon, err))
					return false
				}
				dmgCtx = &details.Effects

				log(fmt.Sprintf("find the best type=%s, target=%s, max_level=%d, weapon=%s", search_type, target, currentLevel, selectedWeapon))
				results, err := steps.GetAllItemsWithTarget(
					api.GetAllItemsFilter{
						Itype:          search_type,
						Craft_material: "",
						Craft_skill:    "",
						Min_level:      strconv.FormatInt(0, 10),
						Max_level:      strconv.FormatInt(int64(currentLevel), 10),
					},
					target,
					dmgCtx,
				)
				if err != nil {
					log(fmt.Sprintf("Failed to get items: %s", err))
					return false
				}

				precision := 3
				for _, res := range (*results)[:min(len(*results), 10)] {
					lvlstr := strconv.FormatInt(int64(res.ItemDetails.Level), 10)
					out := fmt.Sprintf("[%d]", res.ItemDetails.Level) + strings.Repeat(" ", max(0, 3-len(lvlstr)))

					out += res.ItemDetails.Name[:min(len(res.ItemDetails.Name), 24)] + strings.Repeat(" ", max(0, 25-len(res.ItemDetails.Name)))

					totalStr := strconv.FormatFloat(float64(res.HpScore+res.AuxDmgScore+res.AttackScore+res.DmgScore+res.ResistanceScore), 'f', precision, 64)
					out += fmt.Sprintf(" total=%s", totalStr) + strings.Repeat(" ", 6-min(len(totalStr), 6))

					hpScoreStr := strconv.FormatFloat(float64(res.HpScore), 'f', precision, 64)
					out += fmt.Sprintf(" hp=%s", hpScoreStr) + strings.Repeat(" ", 6-min(len(hpScoreStr), 6))

					auxScoreStr := strconv.FormatFloat(float64(res.AuxDmgScore), 'f', precision, 64)
					out += fmt.Sprintf(" aux=%s", auxScoreStr) + strings.Repeat(" ", 6-min(len(auxScoreStr), 6))

					attackScoreStr := strconv.FormatFloat(float64(res.AttackScore), 'f', precision, 64)
					out += fmt.Sprintf(" atk=%s", attackScoreStr) + strings.Repeat(" ", 6-min(len(attackScoreStr), 6))

					dmgScoreStr := strconv.FormatFloat(float64(res.DmgScore), 'f', precision, 64)
					out += fmt.Sprintf(" dmg=%s", dmgScoreStr) + strings.Repeat(" ", 6-min(len(dmgScoreStr), 6))

					resistScoreStr := strconv.FormatFloat(float64(res.ResistanceScore), 'f', precision, 64)
					out += fmt.Sprintf(" resist=%s", resistScoreStr) // + strings.Repeat(" ", 12-max(len(resistScoreStr), 12))

					log(out)
				}

				return true
			}
		} else {
			// original search functionality

			var min_level int64 = 0
			var max_level int64 = 999

			var sorts = make([]steps.SortCri, 0)

			if len(parts) >= 3 {
				maybe_range := parts[2]

				regx_range, err := regexp.Compile(`(\d*)\.\.(\d*)`)
				if err != nil {
					log("failed to compile regexp for level range match")
					return false
				}

				regx_single, err := regexp.Compile(`(\d+)`)
				if err != nil {
					log("failed to compile regexp for level single match")
					return false
				}

				criteria_start := 3
				matches_range := regx_range.FindStringSubmatch(maybe_range)
				matches_single := regx_single.FindStringSubmatch(maybe_range)
				if len(matches_range) > 0 {
					// log("range found")
					maybe_min_level := matches_range[1]
					maybe_max_level := matches_range[2]

					min_level, err = strconv.ParseInt(maybe_min_level, 10, 64)
					if err != nil {
						min_level = 0
					}

					max_level, err = strconv.ParseInt(maybe_max_level, 10, 64)
					if err != nil {
						max_level = 999
					}
				} else if len(matches_single) > 0 {
					maybe_level := matches_single[1]
					// log(fmt.Sprintf("single found: '%s', %v", maybe_level, matches_single))
					min_level, err = strconv.ParseInt(maybe_level, 10, 64)
					if err != nil {
						min_level = 0
					}

					max_level, err = strconv.ParseInt(maybe_level, 10, 64)
					if err != nil {
						max_level = 999
					}
				} else {
					// log("no level")
					criteria_start = 2
				}

				// log(fmt.Sprintf("%d, %v, %d", criteria_start, parts[criteria_start:], len(parts[criteria_start:])))

				for _, cri := range parts[criteria_start:] {
					if cri[0] == '-' {
						sorts = append(sorts, steps.SortCri{
							Equation: []steps.SortEq{
								{
									Prop: cri[1:],
									Op:   "Sub",
								},
							},
						})
					} else {
						sorts = append(sorts, steps.SortCri{
							Equation: []steps.SortEq{
								{
									Prop: cri,
									Op:   "Add",
								},
							},
						})
					}
				}

			}

			sort_str := ""
			if len(sorts) > 0 {
				sort_str = ", sort by "
			}
			for i, c := range sorts {
				if i > 0 {
					sort_str += ", "
				}
				eq := ""
				for i, e := range c.Equation {
					if i == 0 {
						if e.Op == "Sub" {
							eq += "-"
						}
					} else {
						switch e.Op {
						case "Add":
							eq += " + "
						case "Sub":
							eq += " - "
						}
					}

					eq += e.Prop
				}
				sort_str += eq
			}

			log(fmt.Sprintf("find the best type=%s, min_level=%d, max_level=%d%s", search_type, min_level, max_level, sort_str))
			allItems, err := steps.GetAllItemsWithFilter(api.GetAllItemsFilter{
				Itype:          search_type,
				Craft_material: "",
				Craft_skill:    "",
				Min_level:      strconv.FormatInt(min_level, 10),
				Max_level:      strconv.FormatInt(max_level, 10),
			}, sorts)
			if err != nil {
				log(fmt.Sprintf("Failed to get items: %s", err))
				return false
			}

			for _, item := range (*allItems)[:min(len(*allItems), 10)] {
				lvlstr := strconv.FormatInt(int64(item.Level), 10)
				out := fmt.Sprintf("[%d]", item.Level) + strings.Repeat(" ", max(0, 3-len(lvlstr)))

				out += item.Name[:min(len(item.Name), 24)] + strings.Repeat(" ", max(0, 25-len(item.Name)))

				for i, cri := range sorts {
					summation := 0

					for _, eq := range cri.Equation {
						idx := slices.IndexFunc(item.Effects, func(e types.Effect) bool {
							return e.Code == eq.Prop
						})

						if idx >= 0 {
							effect := item.Effects[idx]
							summation += effect.Value
						}
					}

					out += fmt.Sprintf(" eq[%d]=%d", i, summation) + strings.Repeat(" ", max(0, 12))
				}

				log(out)
			}

			return true
		}

	case "simulate-fight":
		if len(parts) != 2 {
			log("usage: simulate-fight <monster_code:string> [[<slot:string>:]<code:string>[:<quantity:int>][,...]]")
			return false
		}

		monster_code := parts[1]

		// TODO: Support custom load-out for fight

		res, err := game.RunSimulations(kernel.CharacterName, monster_code, 1, nil)
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
			log(fmt.Sprintf("Score: %f", (*res)[0].Metadata.Score))
		} else {
			log("no results")
		}

		return true
	case "analyze-fight":
		if len(parts) < 2 || len(parts) > 3 {
			log("usage: analyze-fight <monster_code:string>[ <probability_limit:float64>]")
			return false
		}

		monsterCode := parts[1]
		monsterData, err := api.GetMonsterByCode(monsterCode)
		if err != nil {
			log("failed to get monster info: %s", err)
			return false
		}

		prbLim := 0.0
		if len(parts) >= 3 {
			prbLimRaw := parts[2]
			prbLim, _ = strconv.ParseFloat(prbLimRaw, 64)
		}
		// TODO: Support custom load-out for fight

		characterData := kernel.CharacterState.DeepCopy()

		res, err := game.RunFightAnalysisCore(&characterData, monsterData, nil, prbLim)
		if err != nil {
			log("failed to simulate fight: %s", err)
			return false
		}

		if len(res.EndResults) == 0 {
			log("no results")
			return true
		}

		// log("%v", res)
		turns := make([]int, len(res.EndResults))
		wins, loses := 0, 0
		endHpChar, endHpMonster := 0.0, 0.0
		pWin, pLose := 0.0, 0.0
		for i, r := range res.EndResults {
			if r.CharacterWin {
				wins++
				endHpChar += float64(r.CharacterHp)
				pWin += r.Probability
			} else {
				loses++
				endHpMonster += float64(r.MonsterHp)
				pLose += r.Probability
			}
			turns[i] = r.Turns
		}

		avgTurns := 0.0
		for _, t := range turns {
			avgTurns += float64(t)
		}
		avgTurns /= float64(len(turns))
		endHpChar /= float64(max(1, wins))
		endHpMonster /= float64(max(1, loses))

		log("win")
		log("prb: %f", pWin)
		log("tot: %d", wins)
		log("hp: %f", endHpChar)

		log("lose")
		log("prb: %f", pLose)
		log("tot: %d", loses)
		log("hp: %f", endHpMonster)

		log("turns: %d <-- %f --> %d", slices.Min(turns), avgTurns, slices.Max(turns))
		log("nodes: %d", res.TotalNodes)

		maxHpMonster := monsterData.Hp
		maxHp := 0
		kernel.CharacterState.Read(func(value *types.Character) { maxHp = value.Max_hp })

		colsPerSide := 8
		prbs := make([]float64, colsPerSide*2)
		charPortion := float64(maxHp) / float64(colsPerSide)
		monsterPortion := float64(maxHpMonster) / float64(colsPerSide)

		for _, r := range res.EndResults {
			if r.CharacterWin {
				bucket := int(math.Floor(float64(r.CharacterHp-1) / charPortion))

				// log("char hp: %d", r.CharacterHp)
				// log("bucket: %d", bucket)
				// log("index: %d", colsPerSide+bucket)

				prbs[colsPerSide+bucket] += r.Probability
			} else {
				bucket := int(math.Floor(float64(r.MonsterHp-1) / monsterPortion))
				prbs[colsPerSide-bucket] += r.Probability
			}
		}

		// log("%v", prbs)

		for i, prb := range prbs {
			hp := 0
			if i < colsPerSide {
				// monster
				hp = -(int(float64(maxHpMonster)/float64(colsPerSide)) * (colsPerSide - i))
			} else {
				hp = int(math.Ceil(float64(maxHp)/float64(colsPerSide))) * (i - colsPerSide + 1)
			}

			hpstr := strconv.FormatInt(int64(hp), 10)
			hpout := strings.Repeat(" ", max(0, 5-len(hpstr))) + hpstr

			prbout := strings.Repeat("::", int(math.Ceil(prb/0.1)))

			log("%s - %s", hpout, prbout)
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
				if paused {
					continue
				}

				gen := kernel.Generators[idx]
				if gen == nil {
					continue
				}

				var c = (*gen)(kernel.Last_command, kernel.Last_command_success)
				if c == "noop" {
					kernel.Log(fmt.Sprintf("received noop from generator %d", idx))
					continue
				}

				if c == "clear-gen" {
					c = fmt.Sprintf("clear-gen %d", idx)
				}

				kernel.Commands.With(func(value *[]string) *[]string {
					newValue := append(*value, c)
					return &newValue
				})

				break
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
