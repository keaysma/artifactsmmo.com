package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"artifactsmmo.com/m/gui"
	"artifactsmmo.com/m/utils"
)

var _s = utils.GetSettings()

func print_out(output string) {
	t := time.Now()
	fmt.Printf("[%s] %s %s", t.Format(time.DateTime), s.Character, output)
}

func prompt_line() {
	print_out("< ")
}

func loop_line(content string) {
	// fmt.Println()
	// print_out(fmt.Sprintf("> %s\n", content))
	// prompt_line()

	utils.Log(content) // this is what happens at 3am
}

func Scan_line(scanner *bufio.Scanner) bool {
	prompt_line()
	return scanner.Scan()
}

func __main() {
	if s.Character == "" {
		fmt.Println("Missing env var: character")
		os.Exit(1)
	}

	if s.Api_token == "" {
		fmt.Println("Missing env var: token")
		os.Exit(1)
	}

	go gui.Gameloop()

	scanner := bufio.NewScanner(os.Stdin)

	for Scan_line(scanner) {
		cmd := scanner.Text()
		gui.SharedState.Lock.Lock()
		gui.SharedState.Commands = append(gui.SharedState.Commands, cmd)
		gui.SharedState.Lock.Unlock()

		if cmd == "exit" {
			break
		}
	}

	os.Exit(0)
}
