package gui

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	coords "artifactsmmo.com/m/consts/places"
	generators "artifactsmmo.com/m/game/cmd_generators"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
)

var s = utils.GetSettings()

var running bool

type InternalState struct {
	Current_Generator    generators.Generator
	Last_command         string
	Last_command_success bool
}

var internalState = InternalState{
	Current_Generator:    nil,
	Last_command:         "",
	Last_command_success: false,
}

var reMoveXY = regexp.MustCompile("move (?P<X>[-0-9]+) (?P<Y>[-0-9]+)")
var reDeposit = regexp.MustCompile("deposit (?P<Code>.+) (?P<Quantity>[A-Za-z0-9]+)")
var reEquip = regexp.MustCompile("^equip (?P<Code>.+) (?P<Slot>.+) (?P<Quantity>[0-9]+)")
var reBuy = regexp.MustCompile("^buy (?P<Code>.+) (?P<Quantity>[0-9]+) (?P<MaxPrice>[0-9]+)")
var reSell = regexp.MustCompile("^sell (?P<Code>.+) (?P<Quantity>.+) (?P<MinPrice>[0-9]+)")
var reCraft = regexp.MustCompile("^craft (?P<Code>.+) (?P<Quantity>[0-9]+)")
var reAutoCraft = regexp.MustCompile("auto-craft (?P<Code>.+) (?P<Quantity>[0-9]+)")
var reSleep = regexp.MustCompile("sleep (?P<Time>.+)")

// set gen <GeneratorName> <GeneratorArg> where <GeneratorArg> is optional
var reGenerator = regexp.MustCompile(`gen (?P<GeneratorName>[\w-]+)(?: (?P<GeneratorArg>.+))?`)

func Gameloop() {
	running = true
	for running {

		shared := SharedState.Ref()
		if len(shared.Commands) == 0 {
			SharedState.Unlock()

			if internalState.Current_Generator != nil {
				var c = internalState.Current_Generator(internalState.Last_command, internalState.Last_command_success)
				shared = SharedState.Ref()
				shared.Commands = append(shared.Commands, c)
				SharedState.Unlock()
			}
		} else {
			cmd, commands := shared.Commands[0], shared.Commands[1:]
			shared.Commands = commands
			SharedState.Unlock()

			internalState.Last_command = cmd

			if internalState.Last_command == "ping" {
				fmt.Println()
				utils.Log("pong")
				internalState.Last_command_success = true
			}

			if matches := reSleep.FindStringSubmatch(internalState.Last_command); len(matches) != 0 {
				log := utils.LogPre("sleep: ")

				if sleep_time, err := strconv.ParseFloat(matches[1], 64); err != nil {
					log(fmt.Sprintf("bad time value: %s", matches[1]))
					internalState.Last_command_success = false
				} else {
					time.Sleep(time.Duration(sleep_time * 1_000_000_000))
					log(fmt.Sprintf("slept %f", sleep_time))
					internalState.Last_command_success = true
				}
			}

			matches := reMoveXY.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("move: ")

				x_str := matches[1]
				x, err := strconv.ParseInt(x_str, 10, 64)
				if err != nil {
					log(fmt.Sprintf("can't parse x: %s", x_str))
					internalState.Last_command_success = false
					continue
				}

				y_str := matches[2]
				y, err := strconv.ParseInt(y_str, 10, 64)
				if err != nil {
					log(fmt.Sprintf("can't parse y: %s", y_str))
					internalState.Last_command_success = false
					continue
				}

				_, err = steps.Move(s.Character, coords.Coord{X: int(x), Y: int(y), Name: "<place>"})
				if err != nil {
					log(fmt.Sprintf("failed to move to (%d, %d): %s", x, y, err))
					internalState.Last_command_success = false
				} else {
					internalState.Last_command_success = true
				}
			}

			matches = reDeposit.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("deposit: ")

				code := matches[1]
				quantity_str := matches[2]
				quantity, _ := strconv.ParseInt(quantity_str, 10, 64)
				// if err != nil {
				// 	log(fmt.Sprintf("can't parse quantity: %s", quantity_str))
				// 	internalState.Last_command_success = false
				// }

				_, err := steps.DepositBySelect(
					s.Character,
					func(item types.InventorySlot) bool {
						return item.Code == code
					},
					func(item types.InventorySlot) int {
						if quantity_str == "all" {
							return item.Quantity
						}

						return int(quantity)
					},
				)
				if err != nil {
					log(fmt.Sprintf("failed to deposit %s: %s", code, err))
					internalState.Last_command_success = false
				} else {

					internalState.Last_command_success = true
				}
			}

			matches = reEquip.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("equip: ")

				code := matches[1]
				slot := matches[2]
				quantity_str := matches[3]
				quantity, err := strconv.ParseInt(quantity_str, 10, 64)
				if err != nil {
					log(fmt.Sprintf("can't parse quantity: %s", quantity_str))
					internalState.Last_command_success = false
				}

				err = steps.EquipItem(s.Character, code, slot, int(quantity))
				if err != nil {
					log(fmt.Sprintf("failed to equip %s to %s: %s", code, slot, err))
					internalState.Last_command_success = false
				} else {
					internalState.Last_command_success = true
				}
			}

			matches = reBuy.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("buy: ")

				code := matches[1]
				quantity_str := matches[2]
				quantity, err := strconv.ParseInt(quantity_str, 10, 64)
				if err != nil {
					log(fmt.Sprintf("can't parse quantity: %s", quantity_str))
					internalState.Last_command_success = false
				}
				max_price_str := matches[3]
				max_price, err := strconv.ParseInt(max_price_str, 10, 64)
				if err != nil {
					log(fmt.Sprintf("can't parse max price: %s", max_price_str))
					internalState.Last_command_success = false
				}

				_, err = steps.Buy(s.Character, code, int(quantity), int(max_price))
				if err != nil {
					log(fmt.Sprintf("failed to buy %d %s for %d: %s", quantity, code, max_price, err))
					internalState.Last_command_success = false
				} else {
					internalState.Last_command_success = true
				}
			}

			matches = reSell.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("buselly: ")

				code := matches[1]
				quantity_str := matches[2]
				quantity, _ := strconv.ParseInt(quantity_str, 10, 64) // allow "all"

				min_price_str := matches[3]
				min_price, err := strconv.ParseInt(min_price_str, 10, 64)
				if err != nil {
					log(fmt.Sprintf("can't parse min price: %s", min_price_str))
					internalState.Last_command_success = false
				}

				quantity_func := steps.Amount(int(quantity))
				if quantity_str == "all" {
					quantity_func = func(current_quantity int, max_quantity int) int {
						return min(current_quantity, max_quantity)
					}
				}

				_, err = steps.Sell(s.Character, code, quantity_func, int(min_price))
				if err != nil {
					log(fmt.Sprintf("failed to sell %d %s for %d: %s", quantity, code, min_price, err))
					internalState.Last_command_success = false
				} else {
					internalState.Last_command_success = true
				}
			}

			matches = reCraft.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("craft: ")

				code := matches[1]
				quantity_str := matches[2]
				quantity, err := strconv.ParseInt(quantity_str, 10, 64)
				if err != nil {
					log(fmt.Sprintf("can't parse quantity: %s", quantity_str))
					internalState.Last_command_success = false
				}

				_, err = steps.Craft(s.Character, code, int(quantity))
				if err != nil {
					log(fmt.Sprintf("failed to craft %s: %s", code, err))
					internalState.Last_command_success = false
				} else {
					internalState.Last_command_success = true
				}
			}

			matches = reAutoCraft.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("auto craft: ")

				code := matches[1]
				quantity_str := matches[2]
				quantity, err := strconv.ParseInt(quantity_str, 10, 64)
				if err != nil {
					log(fmt.Sprintf("can't parse quantity: %s", quantity_str))
					internalState.Last_command_success = false
				}

				_, err = steps.AutoCraft(s.Character, code, int(quantity))
				if err != nil {
					log(fmt.Sprintf("failed to craft %s: %s", code, err))
					internalState.Last_command_success = false
				} else {
					internalState.Last_command_success = true
				}
			}

			if internalState.Last_command == "fight" {
				log := utils.LogPre("fight: ")

				_, err := steps.Fight(s.Character, 0)
				if err != nil {
					log(fmt.Sprintf("failed to fight: %s", err))
					internalState.Last_command_success = false
				} else {
					internalState.Last_command_success = true
				}
			}

			if internalState.Last_command == "gather" {
				log := utils.LogPre("gather: ")

				_, err := steps.Gather(s.Character)
				if err != nil {
					log(fmt.Sprintf("failed to gather: %s", err))
					internalState.Last_command_success = false
				} else {
					internalState.Last_command_success = true
				}
			}

			matches = reGenerator.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("set gen: ")

				generator_name := matches[1]

				internalState.Last_command_success = true
				shared_for_gen_name := SharedState.Ref()
				switch generator_name {
				case "make": // special case, handled below
				case "craft-sticky-sword":
					internalState.Current_Generator = generators.Craft_sticky_sword
					shared_for_gen_name.Current_Generator_Name = "craft-sticky-sword"
				case "fight-blue-slimes":
					internalState.Current_Generator = generators.Fight_blue_slimes
					shared_for_gen_name.Current_Generator_Name = "fight-blue-slimes"
				case "ashwood":
					internalState.Current_Generator = generators.Gather_ashwood
					shared_for_gen_name.Current_Generator_Name = "gather-ash-wood"
				case "dummy":
					internalState.Current_Generator = generators.Dummy
					shared_for_gen_name.Current_Generator_Name = "dummy"
				default:
					log(fmt.Sprintf("unknown generator: %s", generator_name))
					internalState.Last_command_success = false
				}
				SharedState.Unlock()

				if generator_name == "make" {
					arg_index := reGenerator.SubexpIndex("GeneratorArg")
					log(fmt.Sprintf("arg_index: %d", arg_index))
					generator_arg := matches[arg_index]

					if generator_arg == "" {
						log("missing generator argument")
						internalState.Last_command_success = false
						continue
					}

					internalState.Current_Generator = generators.Make(generator_arg)
					shared_for_gen_name.Current_Generator_Name = fmt.Sprintf("make <%s>", generator_arg)
				}

				if shared_for_gen_name.Current_Generator_Name != "" {
					log(fmt.Sprintf("generator set to %s", shared_for_gen_name.Current_Generator_Name))
				}
			}

			if internalState.Last_command == "clear gen" {
				internalState.Current_Generator = nil
				SharedState.Ref().Current_Generator_Name = ""
				SharedState.Unlock()
				utils.Log("generator cleared")
				internalState.Last_command_success = true
			}

			if internalState.Last_command == "exit" {
				running = false
			}
		}

		// Nothing happened this loop,
		// Add a small sleep to prevent rapid looping
		time.Sleep(100_000_000) // 100ms (0.1s)
		// time.Sleep(100_000) // 0.1ms
	}
}
