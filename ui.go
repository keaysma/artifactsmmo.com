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

	logs := widgets.NewParagraph()
	logs.Title = "Logs"
	logs.Text = ""

	command_list := widgets.NewParagraph()
	command_list.Title = "Commands"
	command_list.Text = ""

	character_display := widgets.NewTable()
	character_display.Title = s.Character
	character_display.Rows = [][]string{
		{"k", "v"},
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

	loop := func(heavy bool) {
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
		generator_name = shared.Current_Generator_Name
		gui.SharedState.Unlock()

		if generator_name != "" {
			command_list.Title = fmt.Sprintf("Commands (generator: %s)", generator_name)
		} else {
			command_list.Title = "Commands"
		}

		// Updates that run infrequently
		if heavy {
			var gauge_value float64 = 0
			var remaining = time.Duration(0)
			var max = 1
			var now = time.Now()

			cd := state.GlobalCooldown.Ref()
			if cd.End != nil {
				remaining = cd.End.Sub(now)
				max = cd.Duration_seconds
			}
			state.GlobalCooldown.Unlock()

			if remaining.Seconds() < 0 {
				remaining = time.Duration(0)
			}

			gauge_value = (remaining.Seconds() / float64(max))
			cooldown_gauge.Percent = int(gauge_value * 100)

			var character *types.Character = &types.Character{}
			char_state := state.GlobalCharacter.Ref()
			if char_state != nil {
				*character = *char_state
			} else {
				character = nil
			}
			state.GlobalCharacter.Unlock()

			if character != nil {
				character_display.Rows = [][]string{
					{"Name", character.Name},
					{"Position", fmt.Sprintf("(%d, %d)", character.X, character.Y)},
					{"HP", fmt.Sprintf("%d", character.Hp)},
					{"Level", fmt.Sprintf("%d (%d, %d)", character.Level, character.Xp, character.Max_xp)},
					{"Mining", fmt.Sprintf("%d (%d, %d)", character.Mining_level, character.Mining_xp, character.Mining_max_xp)},
					{"Woodcutting", fmt.Sprintf("%d (%d, %d)", character.Woodcutting_level, character.Woodcutting_xp, character.Woodcutting_max_xp)},
				}
			}

		}

		draw(ui.TerminalDimensions())

		// 10 fps, 0.1 seconds
		// time.Sleep(100_000_000)
		// Works fairly well, but a bit slow

		// 20 fps, 0.05 seconds
		time.Sleep(50_000_000)

		// 30 fps, 0.033 seconds
		// time.Sleep(33_333_333)

		// 60 fps, 0.016 seconds
		// time.Sleep(16_000_000)
		// Silky, but the ui keeps flickering
	}

	uiEvents := ui.PollEvents()
	heavy := 0
	for {
		select {
		case event := <-uiEvents:
			switch event.Type {
			case ui.ResizeEvent:
				payload := event.Payload.(ui.Resize)
				draw(payload.Width, payload.Height)
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
		loop(heavy == 0)
		// heavy = (heavy + 1) % 1
	}
}
