package runtimes

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
var command_history = []string{}
var command_history_ptr = 0

var show_tracker = false
var tracker_controls = make(chan string, 2)
var price_updates = make(chan gui.PriceUpdate, 10)
var price_data = [][]float64{
	{1, 1, 1},
	{1, 1, 1},
}

func UI() {
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

	tracker_buy := widgets.NewSparkline()
	tracker_buy.LineColor = ui.ColorGreen
	tracker_sell := widgets.NewSparkline()
	tracker_sell.LineColor = ui.ColorMagenta
	tracker_chart := widgets.NewSparklineGroup(tracker_buy, tracker_sell)
	tracker_chart.Title = "Tracking"
	// tracker_chart.AxesColor = ui.ColorWhite
	// tracker_chart.LineColors[0] = ui.ColorMagenta

	gauge_skill_mining := widgets.NewGauge()
	gauge_skill_woodcutting := widgets.NewGauge()
	gauge_skill_fishing := widgets.NewGauge()
	gauge_skill_weaponcrafting := widgets.NewGauge()
	gauge_skill_gearcrafting := widgets.NewGauge()
	gauge_skill_jewelrycrafting := widgets.NewGauge()
	gauge_skill_cooking := widgets.NewGauge()

	char, err := api.GetCharacterByName(s.Character)
	if err != nil {
		log.Fatalf("failed to get character info for %s: %s", s.Character, err)
		os.Exit(1)
	}
	state.GlobalCharacter.With(func(value *types.Character) *types.Character { return char })

	// If this fails let's just ingore it, not critical
	end, err := time.Parse(time.RFC3339, char.Cooldown_expiration)
	if err == nil {
		state.GlobalCooldown.Set(&state.CooldownData{
			Duration_seconds: char.Cooldown,
			End:              &end,
		})
	}

	go gui.Gameloop()
	go gui.Trackloop(tracker_controls, price_updates)

	draw := func(w int, h int) {
		if show_tracker {
			logs.SetRect(0, 0, w/2, h-24)
			tracker_chart.SetRect(0, h-24, w/2, h-3)
		} else {
			logs.SetRect(0, 0, w/2, h-3)
		}
		command_list.SetRect(w/2, 0, w-(w/4)-1, h-6)
		character_display.SetRect((3*w)/4, 0, w, h-21-6)
		cooldown_gauge.SetRect(w/2, h-6, w, h-3)
		command_entry.SetRect(0, h-3, w, h)

		base_h := h - 21 - 6
		gauge_skill_mining.SetRect((3*w)/4, base_h, w, base_h+3)
		gauge_skill_woodcutting.SetRect((3*w)/4, base_h+3, w, base_h+6)
		gauge_skill_fishing.SetRect((3*w)/4, base_h+6, w, base_h+9)
		gauge_skill_weaponcrafting.SetRect((3*w)/4, base_h+9, w, base_h+12)
		gauge_skill_gearcrafting.SetRect((3*w)/4, base_h+12, w, base_h+15)
		gauge_skill_jewelrycrafting.SetRect((3*w)/4, base_h+15, w, base_h+18)
		gauge_skill_cooking.SetRect((3*w)/4, base_h+18, w, base_h+21)

		ui.Render(
			logs, command_list, command_entry, character_display, cooldown_gauge,
			gauge_skill_mining, gauge_skill_woodcutting, gauge_skill_fishing, gauge_skill_weaponcrafting, gauge_skill_gearcrafting, gauge_skill_jewelrycrafting, gauge_skill_cooking,
		)
		if show_tracker {
			ui.Render(tracker_chart)
		}
	}

	draw(ui.TerminalDimensions())

	loop := func(heavy bool) {
		select {
		case line := <-utils.LogsChannel:
			log_lines = append(log_lines, line)
		case price_update := <-price_updates:
			price_data[0] = append(price_data[0], float64(price_update.Buy))
			price_data[1] = append(price_data[1], float64(price_update.Sell))
		default:
		}
		h := logs.Inner.Dy()
		if len(log_lines) > h {
			log_lines = log_lines[max(0, len(log_lines)-h):]
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
			var max_dur = 1
			var now = time.Now()

			cd := state.GlobalCooldown.Ref()
			if cd.End != nil {
				remaining = cd.End.Sub(now)
				max_dur = cd.Duration_seconds
			}
			state.GlobalCooldown.Unlock()

			if remaining.Seconds() < 0 {
				remaining = time.Duration(0)
			}

			gauge_value = (remaining.Seconds() / float64(max_dur))
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
					{"Position", fmt.Sprintf("(%d, %d)", character.X, character.Y)},
					{"HP", fmt.Sprintf("%d", character.Hp)},
					{"Level", fmt.Sprintf("%d %d/%d", character.Level, character.Xp, character.Max_xp)},
					{"Task", fmt.Sprintf("%s %d/%d", character.Task, character.Task_progress, character.Task_total)},
					{"Gold", fmt.Sprintf("%d", character.Gold)},
					// {"Mining", fmt.Sprintf("%d %d/%d", character.Mining_level, character.Mining_xp, character.Mining_max_xp)},
					// {"Woodcutting", fmt.Sprintf("%d %d/%d", character.Woodcutting_level, character.Woodcutting_xp, character.Woodcutting_max_xp)},
				}

				gauge_skill_mining.Title = fmt.Sprintf("Mining: %d", character.Mining_level)
				gauge_skill_mining.Percent = int((float64(character.Mining_xp) / float64(character.Mining_max_xp)) * 100)

				gauge_skill_woodcutting.Title = fmt.Sprintf("Woodcutting: %d", character.Woodcutting_level)
				gauge_skill_woodcutting.Percent = int((float64(character.Woodcutting_xp) / float64(character.Woodcutting_max_xp)) * 100)

				gauge_skill_fishing.Title = fmt.Sprintf("Fishing: %d", character.Fishing_level)
				gauge_skill_fishing.Percent = int((float64(character.Fishing_xp) / float64(character.Fishing_max_xp)) * 100)

				gauge_skill_weaponcrafting.Title = fmt.Sprintf("Weapon Crafting: %d", character.Weaponcrafting_level)
				gauge_skill_weaponcrafting.Percent = int((float64(character.Weaponcrafting_xp) / float64(character.Weaponcrafting_max_xp)) * 100)

				gauge_skill_gearcrafting.Title = fmt.Sprintf("Gear Crafting: %d", character.Gearcrafting_level)
				gauge_skill_gearcrafting.Percent = int((float64(character.Gearcrafting_xp) / float64(character.Gearcrafting_max_xp)) * 100)

				gauge_skill_jewelrycrafting.Title = fmt.Sprintf("Jewelry Crafting: %d", character.Jewelrycrafting_level)
				gauge_skill_jewelrycrafting.Percent = int((float64(character.Jewelrycrafting_xp) / float64(character.Jewelrycrafting_max_xp)) * 100)

				gauge_skill_cooking.Title = fmt.Sprintf("Cooking: %d", character.Cooking_level)
				gauge_skill_cooking.Percent = int((float64(character.Cooking_xp) / float64(character.Cooking_max_xp)) * 100)
			}

			tracker_buy.Data = price_data[0][max(0, len(price_data[0])-tracker_chart.Inner.Dx()):]
			tracker_sell.Data = price_data[1][max(0, len(price_data[1])-tracker_chart.Inner.Dx()):]

			if len(price_data[0]) > 0 {
				tracker_buy.Title = fmt.Sprintf("buy: %d", int(price_data[0][len(price_data[0])-1]))
			}
			if len(price_data[1]) > 0 {
				tracker_sell.Title = fmt.Sprintf("sell: %d", int(price_data[1][len(price_data[1])-1]))
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
				case "<C-c>", "<C-v>":
				case "<Left>", "<Right>":
				case "<Up>":
					if command_history_ptr >= len(command_history) {
						command_history_ptr = 0
					}
					if len(command_history) != 0 {
						command_history_ptr += 1

						command_value = command_history[len(command_history)-command_history_ptr]
					}
				case "<Down>":
					if command_history_ptr <= 0 {
						command_history_ptr = len(command_history)
					}
					if len(command_history) != 0 {
						command_history_ptr -= 1

						command_value = command_history[len(command_history)-command_history_ptr-1]
					}
				case "<Enter>":
					if command_value == "exit" {
						return // bye!
					} else if command_value == "help" {
						// utils.Log pushes to log channel and that deadlocks
						log_lines = append(log_lines, "help message")
					} else if command_value == "clear" {
						log_lines = []string{}
					} else if command_value == "stop" {
						gui.SharedState.With(func(value *gui.SharedStateType) *gui.SharedStateType {
							value.Commands = []string{}
							return value
						})
					} else if strings.HasPrefix(command_value, "track") {
						parts := strings.Split(command_value, " ")
						price_data = [][]float64{
							{},
							{},
						}

						stop_tracking := len(parts) != 2

						show_tracker = !stop_tracking
						if show_tracker {
							tracker_chart.Title = fmt.Sprintf("Tracking: %s", parts[1])
							tracker_controls <- parts[1]
						} else {
							tracker_controls <- ""
						}

					} else if command_value != "" {
						command_history = append(command_history[max(0, len(command_history)-50):], command_value)
						shared := gui.SharedState.Ref()
						shared.Commands = append(shared.Commands, command_value)
						gui.SharedState.Unlock()
					}
					command_value = ""
					command_history_ptr = 0
				case "<Backspace>", "<C-<Backspace>>":
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
		heavy = (heavy + 1) % 6
	}
}
