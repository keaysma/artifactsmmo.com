package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/gui"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var s = utils.GetSettings()

var command_value = ""
var log_lines = []string{}

func main() {
	err := ui.Init()
	if err != nil {
		log.Fatalf("failed to initialize termui: %s", err)
	}

	defer ui.Close()

	w, h := ui.TerminalDimensions()

	logs := widgets.NewParagraph()
	logs.Title = "Logs"
	logs.Text = ""

	command_list := widgets.NewParagraph()
	command_list.Title = "Commands"
	command_list.Text = ""

	character_display := widgets.NewTable()
	character_display.Title = s.Character
	character_display.Rows = [][]string{
		[]string{"k", "v"},
		[]string{"hp", "0"},
	}

	cooldown_gauge := widgets.NewGauge()
	cooldown_gauge.Title = "Cooldown"

	command_entry := widgets.NewParagraph()
	command_entry.Text = "> "
	command_entry.BorderStyle.Fg = ui.ColorBlue

	char, err := api.GetCharacterByName(s.Character)
	if err != nil {
		log.Fatalf("failed to get character info for %s: %s", s.Character, err)
		os.Exit(1)
	}
	state.GlobalCharacter.With(func(value *types.Character) *types.Character { return char })

	go gui.Gameloop()

	draw := func(w int, h int) {
		logs.SetRect(0, 0, w/2, h-3)
		command_list.SetRect(w/2, 0, w, h-6)
		character_display.SetRect((3*w)/4, h-20, w-1, h-7)
		cooldown_gauge.SetRect(w/2, h-6, w, h-3)
		command_entry.SetRect(0, h-3, w, h)
		ui.Render(logs, command_list, command_entry, character_display, cooldown_gauge)
	}

	draw(ui.TerminalDimensions())

	loop := func() {
		select {
		case line := <-utils.LogsChannel:
			log_lines = append(log_lines, line)
		default:
		}
		if len(log_lines) > 50 {
			log_lines = log_lines[len(log_lines)-50:]
		}
		logs.Text = strings.Join(log_lines, "\n")

		generator_name := ""
		shared := gui.SharedState.Ref()
		command_list.Text = strings.Join(shared.Commands, "\n")

		var character *types.Character
		char_state := state.GlobalCharacter.Ref()
		if char_state != nil {
			character = &(*char_state)
		}
		state.GlobalCharacter.Unlock()

		generator_name = shared.Current_Generator_Name
		gui.SharedState.Unlock()

		if character != nil {
			character_display.Rows = [][]string{
				[]string{"Name", character.Name},
				[]string{"Position", fmt.Sprintf("(%d, %d)", character.X, character.Y)},
				[]string{"HP", fmt.Sprintf("%d", character.Hp)},
				[]string{"Level", fmt.Sprintf("%d (%d, %d)", character.Level, character.Xp, character.Max_xp)},
				[]string{"Mining", fmt.Sprintf("%d (%d, %d)", character.Mining_level, character.Mining_xp, character.Mining_max_xp)},
				[]string{"Woodcutting", fmt.Sprintf("%d (%d, %d)", character.Woodcutting_level, character.Woodcutting_xp, character.Woodcutting_max_xp)},
			}
		}

		if generator_name != "" {
			command_list.Title = fmt.Sprintf("Commands (generator: %s)", generator_name)
		} else {
			command_list.Title = "Commands"
		}

		var gauge_value float64 = 0
		cd := state.GlobalCooldown.Ref()
		if cd != nil && cd.Current > 0 {
			(*cd).Current = max((*cd).Current-float64(1)/10, 0)
		}
		gauge_value = (cd.Current / cd.Max)
		state.GlobalCooldown.Unlock()
		cooldown_gauge.Percent = int(gauge_value * 100)

		draw(ui.TerminalDimensions())

		time.Sleep(1_000_000_00)
	}

	uiEvents := ui.PollEvents()
	for {
		select {
		case event := <-uiEvents:
			switch event.Type {
			case ui.ResizeEvent:
				payload := event.Payload.(ui.Resize)
				w, h = payload.Width, payload.Height
				draw(w, h)
			case ui.KeyboardEvent:
				switch event.ID {
				// no-ops
				case "<Escape>":
				case "<C-c>":
				case "<C-v>":
				case "<Enter>":
					if command_value == "exit" {
						return
					} else if command_value == "help" {
						// utils.Log("help message")
						// utils.Log pushes to log channel and that deadlocks
						// probably need to push directly to log_lines?
					} else {
						shared := gui.SharedState.Ref()
						shared.Commands = append(shared.Commands, command_value)
						gui.SharedState.Unlock()
					}
					command_value = ""
				case "<Backspace>":
					if len(command_value) > 0 {
						command_value = command_value[:len(command_value)-1]
					}
				case "<Space>":
					if len(command_value) > 0 {
						command_value += " "
					}
				default:
					command_value += event.ID

				}
				command_entry.Text = fmt.Sprintf("> %s", command_value)

			}
		default:
		}
		loop()
	}
}
