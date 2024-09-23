package main

import (
	"bufio"
	"fmt"
	"os"

	"artifactsmmo.com/m/gui"
	"artifactsmmo.com/m/utils"
)

var s = utils.GetSettings()

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

	for gui.Scan_line(scanner) {
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
