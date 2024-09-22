package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	coords "artifactsmmo.com/m/consts/places"
	"artifactsmmo.com/m/game/steps"
	"artifactsmmo.com/m/utils"
)

var s = utils.GetSettings()

func print_out(output string) {
	t := time.Now()
	fmt.Printf("[%s] %s %s", t.Format(time.DateTime), s.Character, output)
}

func prompt_line() {
	print_out("< ")
}

func loop_line(content string) {
	fmt.Println()
	print_out(fmt.Sprintf("> %s\n", content))
	prompt_line()
}

func scan_line(scanner *bufio.Scanner) bool {
	prompt_line()
	return scanner.Scan()
}

func dummy_generator(last string) string {
	if last == "sleep 1" {
		return "ping"
	}

	return "sleep 1"
}

var running bool

type SharedState struct {
	Lock     sync.Mutex
	Commands []string
}

var sharedState SharedState = SharedState{
	Lock:     sync.Mutex{},
	Commands: []string{},
}

type Generator func(ctx string) string
type InternalState struct {
	Current_Generator Generator
	Last_command      string
}

var internalState InternalState = InternalState{
	Current_Generator: nil,
	Last_command:      "",
}

var reMoveXY = regexp.MustCompile("move (?P<X>[-0-9]+) (?P<Y>[-0-9]+)")
var reSleep = regexp.MustCompile("sleep (?P<Time>.+)")
var reGenerator = regexp.MustCompile("set gen (?P<GeneratorName>.+)")

func gameloop() {
	running = true
	for running {

		sharedState.Lock.Lock()
		if len(sharedState.Commands) == 0 {
			sharedState.Lock.Unlock()

			if internalState.Current_Generator != nil {
				var c = internalState.Current_Generator(internalState.Last_command)
				sharedState.Lock.Lock()
				sharedState.Commands = append(sharedState.Commands, c)
				sharedState.Lock.Unlock()
			}
		} else {
			cmd, commands := sharedState.Commands[0], sharedState.Commands[1:]
			sharedState.Commands = commands
			sharedState.Lock.Unlock()
			internalState.Last_command = cmd

			if internalState.Last_command == "ping" {
				fmt.Println()
				loop_line("pong")
			}

			matches := reSleep.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				timeIndex := reSleep.SubexpIndex("Time")
				sleep_time_str := matches[timeIndex]
				sleep_time, err := strconv.ParseFloat(sleep_time_str, 64)
				if err != nil {
					loop_line(fmt.Sprintf("bad time value: %s", sleep_time_str))
				} else {
					time.Sleep(time.Duration(sleep_time * 1_000_000_000))
					loop_line(fmt.Sprintf("slept %f", sleep_time))
				}
			}

			matches = reMoveXY.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				x_index := reMoveXY.SubexpIndex("X")
				if x_index < 0 {
					loop_line("can't read x")
					continue
				}
				x_str := matches[x_index]
				x, err := strconv.ParseInt(x_str, 10, 64)
				if err != nil {
					loop_line(fmt.Sprintf("failed to read x: %s", x_str))
					continue
				}

				y_index := reMoveXY.SubexpIndex("Y")
				if y_index < 0 {
					loop_line("can't read y")
					continue
				}
				y_str := matches[y_index]
				y, err := strconv.ParseInt(y_str, 10, 64)
				if err != nil {
					loop_line(fmt.Sprintf("failed to read y: %s", y_str))
					continue
				}

				loop_line(fmt.Sprintf("moving to (%d, %d)", x, y))
				err = steps.Move(s.Character, coords.Coord{X: int(x), Y: int(y), Name: "<place>"})
				if err != nil {
					loop_line(fmt.Sprintf("moved to (%d, %d)", x, y))
				} else {
					loop_line(fmt.Sprintf("failed to move to (%d, %d): %s", x, y, err))
				}
			}

			if internalState.Last_command == "fight" {
				loop_line("fighting!")
				_, err := steps.Fight(s.Character, 0)
				if err == nil {
					loop_line("done fighting")
				} else {
					loop_line(fmt.Sprintf("failed to fight: %s", err))
				}
			}

			if internalState.Last_command == "gather" {
				loop_line("gathering!")
				_, err := steps.Gather(s.Character)
				if err == nil {
					loop_line("done gathering")
				} else {
					loop_line(fmt.Sprintf("failed to gather: %s", err))
				}
			}

			matches = reGenerator.FindStringSubmatch(internalState.Last_command)
			if len(matches) != 0 {
				generator_name_index := reGenerator.SubexpIndex("GeneratorName")
				generator_name := matches[generator_name_index]

				if generator_name == "dummy" {
					internalState.Current_Generator = dummy_generator
					loop_line("generator set to dummy")
				} else {
					loop_line("unknown generator")
				}
			}

			if internalState.Last_command == "clear gen" {
				internalState.Current_Generator = nil
				loop_line("generator cleared")
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

func main() {
	if s.Character == "" {
		fmt.Println("Missing env var: character")
		os.Exit(1)
	}

	if s.Api_token == "" {
		fmt.Println("Missing env var: token")
		os.Exit(1)
	}

	go gameloop()

	scanner := bufio.NewScanner(os.Stdin)

	for scan_line(scanner) {
		cmd := scanner.Text()
		sharedState.Lock.Lock()
		sharedState.Commands = append(sharedState.Commands, cmd)
		sharedState.Lock.Unlock()

		if cmd == "exit" {
			break
		}
	}

	os.Exit(0)
}
