package gui

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/utils"
)

var s = utils.GetSettings()

var running bool

type Generator func(ctx string, success bool) string
type InternalState struct {
	Current_Generator    Generator
	Last_command         string
	Last_command_success bool
}

var internalState = InternalState{
	Current_Generator:    nil,
	Last_command:         "",
	Last_command_success: false,
}

var reMoveXY = regexp.MustCompile("move (?P<X>[-0-9]+) (?P<Y>[-0-9]+)")
var reCraft = regexp.MustCompile("^craft (?P<Code>.+) (?P<Quantity>[0-9]+)")
var reAutoCraft = regexp.MustCompile("auto-craft (?P<Code>.+) (?P<Quantity>[0-9]+)")
var reSleep = regexp.MustCompile("sleep (?P<Time>.+)")
var reGenerator = regexp.MustCompile("set gen (?P<GeneratorName>.+)")

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

			matches := reSleep.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("sleep: ")

				sleep_time_str := matches[1]
				sleep_time, err := strconv.ParseFloat(sleep_time_str, 64)
				if err != nil {
					log(fmt.Sprintf("bad time value: %s", sleep_time_str))
					internalState.Last_command_success = false
				} else {
					time.Sleep(time.Duration(sleep_time * 1_000_000_000))
					log(fmt.Sprintf("slept %f", sleep_time))
					internalState.Last_command_success = true
				}
			}

			matches = reMoveXY.FindStringSubmatch(internalState.Last_command)
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
				case "fight blue slimes":
					internalState.Current_Generator = game.Fight_blue_slimes
					shared_for_gen_name.Current_Generator_Name = "fight blue slimes"
				case "ashwood":
					internalState.Current_Generator = game.Gather_ashwood
					shared_for_gen_name.Current_Generator_Name = "gather ash wood"
				case "dummy":
					internalState.Current_Generator = game.Dummy
					shared_for_gen_name.Current_Generator_Name = "dummy"
				default:
					log(fmt.Sprintf("unknown generator: %s", generator_name))
					internalState.Last_command_success = false
				}

				if shared_for_gen_name.Current_Generator_Name != "" {
					log(fmt.Sprintf("generator set to %s", shared_for_gen_name.Current_Generator_Name))
				}
				SharedState.Unlock()
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
