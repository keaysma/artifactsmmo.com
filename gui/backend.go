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

type Generator func(ctx string) string
type InternalState struct {
	Current_Generator Generator
	Last_command      string
}

var internalState = InternalState{
	Current_Generator: nil,
	Last_command:      "",
}

var reMoveXY = regexp.MustCompile("move (?P<X>[-0-9]+) (?P<Y>[-0-9]+)")
var reSleep = regexp.MustCompile("sleep (?P<Time>.+)")
var reGenerator = regexp.MustCompile("set gen (?P<GeneratorName>.+)")

func Gameloop() {
	running = true
	for running {

		shared := SharedState.Ref()
		if len(shared.Commands) == 0 {
			SharedState.Unlock()

			if internalState.Current_Generator != nil {
				var c = internalState.Current_Generator(internalState.Last_command)
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
			}

			matches := reSleep.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("sleep: ")

				timeIndex := reSleep.SubexpIndex("Time")
				sleep_time_str := matches[timeIndex]
				sleep_time, err := strconv.ParseFloat(sleep_time_str, 64)
				if err != nil {
					log(fmt.Sprintf("bad time value: %s", sleep_time_str))
				} else {
					time.Sleep(time.Duration(sleep_time * 1_000_000_000))
					log(fmt.Sprintf("slept %f", sleep_time))
				}
			}

			matches = reMoveXY.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("move: ")

				x_index := reMoveXY.SubexpIndex("X")
				if x_index < 0 {
					log("can't match x")
					continue
				}
				x_str := matches[x_index]
				x, err := strconv.ParseInt(x_str, 10, 64)
				if err != nil {
					log(fmt.Sprintf("can't parse x: %s", x_str))
					continue
				}

				y_index := reMoveXY.SubexpIndex("Y")
				if y_index < 0 {
					log("can't match y")
					continue
				}
				y_str := matches[y_index]
				y, err := strconv.ParseInt(y_str, 10, 64)
				if err != nil {
					log(fmt.Sprintf("can't parse y: %s", y_str))
					continue
				}

				_, err = steps.Move(s.Character, coords.Coord{X: int(x), Y: int(y), Name: "<place>"})
				if err != nil {
					log(fmt.Sprintf("failed to move to (%d, %d): %s", x, y, err))
				}
			}

			if internalState.Last_command == "fight" {
				log := utils.LogPre("fight: ")

				_, err := steps.Fight(s.Character, 0)
				if err != nil {
					log(fmt.Sprintf("failed to fight: %s", err))
				}
			}

			if internalState.Last_command == "gather" {
				log := utils.LogPre("gather: ")

				_, err := steps.Gather(s.Character)
				if err != nil {
					log(fmt.Sprintf("failed to gather: %s", err))
				}
			}

			matches = reGenerator.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				log := utils.LogPre("set gen: ")

				generator_name_index := reGenerator.SubexpIndex("GeneratorName")
				generator_name := matches[generator_name_index]

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
