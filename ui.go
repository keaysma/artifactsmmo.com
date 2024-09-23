package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"artifactsmmo.com/m/gui"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/utils"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var command_value = ""

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
	logs.SetRect(0, 0, w/2, h-3)
	ui.Render(logs)

	command_list := widgets.NewParagraph()
	command_list.Title = "Commands"
	command_list.Text = ""
	command_list.SetRect(w/2, 0, w, h-6)
	ui.Render(command_list)

	cooldown_gauge := widgets.NewGauge()
	cooldown_gauge.Title = "Cooldown"
	cooldown_gauge.SetRect(w/2, h-6, w, h-3)
	ui.Render(cooldown_gauge)

	command_entry := widgets.NewParagraph()
	command_entry.Text = "> "
	command_entry.BorderStyle.Fg = ui.ColorBlue
	command_entry.SetRect(0, h-3, w, h)
	ui.Render(command_entry)

	go gui.Gameloop()

	draw := func() {
		logs.Text = utils.LogsAsString()

		generator_name := ""
		gui.SharedState.Lock.Lock()
		command_list.Text = strings.Join(gui.SharedState.Commands, "\n")
		generator_name = gui.SharedState.Current_Generator_Name
		gui.SharedState.Lock.Unlock()

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

		ui.Render(logs, command_list, command_entry, cooldown_gauge)

		time.Sleep(1_000_000_00)
	}

	uiEvents := ui.PollEvents()
	for {
		select {
		case event := <-uiEvents:
			switch event.Type {
			case ui.ResizeEvent:
				payload := event.Payload.(ui.Resize)
				logs.SetRect(0, 0, payload.Width/2, payload.Height-3)
				command_list.SetRect(payload.Width/2, 0, payload.Width, payload.Height-6)
				cooldown_gauge.SetRect(payload.Width/2, payload.Height-6, payload.Width, payload.Height-3)
				command_entry.SetRect(0, payload.Height-3, payload.Width, payload.Height)
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
						utils.Log("help message")
					} else {
						gui.SharedState.Lock.Lock()
						gui.SharedState.Commands = append(gui.SharedState.Commands, command_value)
						gui.SharedState.Lock.Unlock()
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
		draw()
	}
}
