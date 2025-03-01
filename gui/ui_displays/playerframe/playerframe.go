package playerframe

// gonna move all the mainframe widgets + loop logic here
// do the same thing for amm after in another file

import (
	"fmt"
	"os"
	"strings"
	"time"

	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/utils"
	ui "github.com/keaysma/termui/v3"
	"github.com/keaysma/termui/v3/widgets"
)

var commandValue = ""
var logLines = []string{}
var commandHistory = []string{}
var commandHistory_ptr = 0

type Mainframe struct {
	kernel                    *game.Kernel
	Logs                      *widgets.Paragraph
	CommandList               *widgets.Paragraph
	OrderReferenceList        *widgets.Paragraph
	CharacterDisplay          *widgets.Table
	CooldownGauge             *widgets.Gauge
	CommandEntry              *widgets.Paragraph
	GaugeSkillMining          *widgets.Gauge
	GaugeSkillWoodcutting     *widgets.Gauge
	GaugeSkillFishing         *widgets.Gauge
	GaugeSkillWeaponcrafting  *widgets.Gauge
	GaugeSkillGearcrafting    *widgets.Gauge
	GaugeSkillJewelrycrafting *widgets.Gauge
	GaugeSkillCooking         *widgets.Gauge

	// Settings
	TabHeight int
}

func Init(s *utils.Settings, kernel *game.Kernel) *Mainframe {
	logs := widgets.NewParagraph()
	logs.Title = "Logs"
	logs.Text = ""

	commandList := widgets.NewParagraph()
	commandList.Title = "Commands"
	commandList.Text = ""

	orderReferenceList := widgets.NewParagraph()
	orderReferenceList.Title = "Order Reference"
	orderReferenceList.Text = ""

	characterDisplay := widgets.NewTable()
	characterDisplay.Title = kernel.CharacterName
	characterDisplay.Rows = [][]string{
		{"k", "v"},
	}

	cooldownGauge := widgets.NewGauge()
	cooldownGauge.Title = "Cooldown"

	command_entry := widgets.NewParagraph()
	command_entry.Text = "> "
	command_entry.BorderStyle.Fg = ui.ColorBlue

	gauge_skill_mining := widgets.NewGauge()
	gauge_skill_woodcutting := widgets.NewGauge()
	gauge_skill_fishing := widgets.NewGauge()
	gauge_skill_weaponcrafting := widgets.NewGauge()
	gauge_skill_gearcrafting := widgets.NewGauge()
	gauge_skill_jewelrycrafting := widgets.NewGauge()
	gauge_skill_cooking := widgets.NewGauge()

	mainframeWigets := Mainframe{
		kernel:                    kernel,
		Logs:                      logs,
		CommandList:               commandList,
		OrderReferenceList:        orderReferenceList,
		CharacterDisplay:          characterDisplay,
		CooldownGauge:             cooldownGauge,
		CommandEntry:              command_entry,
		GaugeSkillMining:          gauge_skill_mining,
		GaugeSkillWoodcutting:     gauge_skill_woodcutting,
		GaugeSkillFishing:         gauge_skill_fishing,
		GaugeSkillWeaponcrafting:  gauge_skill_weaponcrafting,
		GaugeSkillGearcrafting:    gauge_skill_gearcrafting,
		GaugeSkillJewelrycrafting: gauge_skill_jewelrycrafting,
		GaugeSkillCooking:         gauge_skill_cooking,

		TabHeight: s.TabHeight,
	}

	return &mainframeWigets
}

func (m *Mainframe) Draw() {
	ui.Render(m.Logs, m.CommandList, m.CharacterDisplay, m.CooldownGauge, m.CommandEntry, m.GaugeSkillMining, m.GaugeSkillWoodcutting, m.GaugeSkillFishing, m.GaugeSkillWeaponcrafting, m.GaugeSkillGearcrafting, m.GaugeSkillJewelrycrafting, m.GaugeSkillCooking)

	if m.OrderReferenceList.Text != "" {
		ui.Render(m.OrderReferenceList)
	}
}

func (m *Mainframe) ResizeWidgets(w int, h int) {
	m.Logs.SetRect(0, m.TabHeight, w/2, h-3)
	m.CommandList.SetRect(w/2, m.TabHeight, w-(w/4)-1, h-6)
	m.OrderReferenceList.SetRect(w/2, h-18, w-(w/4)-1, h-6)
	m.CharacterDisplay.SetRect((3*w)/4, m.TabHeight, w, h-21-6)
	m.CooldownGauge.SetRect(w/2, h-6, w, h-3)
	m.CommandEntry.SetRect(0, h-3, w, h)

	base_h := h - 21 - 6
	m.GaugeSkillMining.SetRect((3*w)/4, base_h, w, base_h+3)
	m.GaugeSkillWoodcutting.SetRect((3*w)/4, base_h+3, w, base_h+6)
	m.GaugeSkillFishing.SetRect((3*w)/4, base_h+6, w, base_h+9)
	m.GaugeSkillWeaponcrafting.SetRect((3*w)/4, base_h+9, w, base_h+12)
	m.GaugeSkillGearcrafting.SetRect((3*w)/4, base_h+12, w, base_h+15)
	m.GaugeSkillJewelrycrafting.SetRect((3*w)/4, base_h+15, w, base_h+18)
	m.GaugeSkillCooking.SetRect((3*w)/4, base_h+18, w, base_h+21)
}

func (m *Mainframe) Loop(heavy bool) {
	select {
	case line := <-utils.LogsChannel:
		logLines = append(logLines, line)
	default:
	}
	h := m.Logs.Inner.Dy()
	if len(logLines) > h {
		logLines = logLines[max(0, len(logLines)-h):]
	}
	m.Logs.Text = strings.Join(logLines, "\n")

	m.CommandList.Text = strings.Join(m.kernel.Commands.ShallowCopy(), "\n")

	generator_name := m.kernel.CurrentGeneratorName.ShallowCopy()
	if generator_name != "" {
		m.CommandList.Title = fmt.Sprintf("Commands (generator: %s)", generator_name)
	} else {
		m.CommandList.Title = "Commands"
	}

	// Updates that run infrequently
	if heavy {
		var gauge_value float64 = 0
		var remaining = time.Duration(0)
		var max_dur = 1
		var now = time.Now()

		m.kernel.CooldownState.With(func(value *state.CooldownData) *state.CooldownData {
			if value.End != nil {
				remaining = value.End.Sub(now)
				max_dur = value.Duration_seconds
			}
			return value
		})

		if remaining.Seconds() < 0 {
			remaining = time.Duration(0)
		}

		gauge_value = (remaining.Seconds() / float64(max_dur))
		m.CooldownGauge.Percent = int(gauge_value * 100)

		character := m.kernel.CharacterState.Ref()

		// if character != nil {
		m.CharacterDisplay.Rows = [][]string{
			{"Position", fmt.Sprintf("(%d, %d)", character.X, character.Y)},
			{"HP", fmt.Sprintf("%d/%d", character.Hp, character.Max_hp)},
			{"Level", fmt.Sprintf("%d %d/%d", character.Level, character.Xp, character.Max_xp)},
			{"Task", fmt.Sprintf("%s %d/%d", character.Task, character.Task_progress, character.Task_total)},
			{"Gold", fmt.Sprintf("%d", character.Gold)},
		}

		m.GaugeSkillMining.Title = fmt.Sprintf("Mining: %d", character.Mining_level)
		m.GaugeSkillMining.Percent = int((float64(character.Mining_xp) / float64(character.Mining_max_xp)) * 100)

		m.GaugeSkillWoodcutting.Title = fmt.Sprintf("Woodcutting: %d", character.Woodcutting_level)
		m.GaugeSkillWoodcutting.Percent = int((float64(character.Woodcutting_xp) / float64(character.Woodcutting_max_xp)) * 100)

		m.GaugeSkillFishing.Title = fmt.Sprintf("Fishing: %d", character.Fishing_level)
		m.GaugeSkillFishing.Percent = int((float64(character.Fishing_xp) / float64(character.Fishing_max_xp)) * 100)

		m.GaugeSkillWeaponcrafting.Title = fmt.Sprintf("Weapon Crafting: %d", character.Weaponcrafting_level)
		m.GaugeSkillWeaponcrafting.Percent = int((float64(character.Weaponcrafting_xp) / float64(character.Weaponcrafting_max_xp)) * 100)

		m.GaugeSkillGearcrafting.Title = fmt.Sprintf("Gear Crafting: %d", character.Gearcrafting_level)
		m.GaugeSkillGearcrafting.Percent = int((float64(character.Gearcrafting_xp) / float64(character.Gearcrafting_max_xp)) * 100)

		m.GaugeSkillJewelrycrafting.Title = fmt.Sprintf("Jewelry Crafting: %d", character.Jewelrycrafting_level)
		m.GaugeSkillJewelrycrafting.Percent = int((float64(character.Jewelrycrafting_xp) / float64(character.Jewelrycrafting_max_xp)) * 100)

		m.GaugeSkillCooking.Title = fmt.Sprintf("Cooking: %d", character.Cooking_level)
		m.GaugeSkillCooking.Percent = int((float64(character.Cooking_xp) / float64(character.Cooking_max_xp)) * 100)
		// }

		m.kernel.CharacterState.Unlock()

		m.OrderReferenceList.Text = ""
		ordersList := state.OrderIdsReference.Ref()
		for i, id := range *ordersList {
			m.OrderReferenceList.Text += fmt.Sprintf("%d: %s\n", i, id)
		}
		state.OrderIdsReference.Unlock()
	}
}

var PRIORITY_COMMANDS = []string{"o", "myo", "simulate-fight"}

func (m *Mainframe) HandleKeyboardInput(event ui.Event) {
	switch event.ID {
	case "<Up>":
		if commandHistory_ptr >= len(commandHistory) {
			commandHistory_ptr = 0
		}
		if len(commandHistory) != 0 {
			commandHistory_ptr += 1

			commandValue = commandHistory[len(commandHistory)-commandHistory_ptr]
		}
	case "<Down>":
		if commandHistory_ptr <= 0 {
			commandHistory_ptr = len(commandHistory)
		}
		if len(commandHistory) != 0 {
			commandHistory_ptr -= 1

			commandValue = commandHistory[len(commandHistory)-commandHistory_ptr-1]
		}
	case "<Enter>":
		if commandValue == "exit" {
			// bye!
			ui.Close()
			os.Exit(0)
		} else if commandValue == "help" {
			// utils.Log pushes to log channel and that deadlocks
			// logLines = append(logLines, "help message")
			utils.Log("help message")
		} else if commandValue == "clear" {
			logLines = []string{}
			state.OrderIdsReference.Set(&[]string{})
		} else if commandValue == "stop" {
			m.kernel.Commands.Set(&[]string{})
		} else if utils.Contains(PRIORITY_COMMANDS, strings.Split(commandValue, " ")[0]) {
			commandHistory = append(commandHistory[max(0, len(commandHistory)-50):], commandValue)
			m.kernel.PriorityCommands <- commandValue
		} else if commandValue != "" {
			commandHistory = append(commandHistory[max(0, len(commandHistory)-50):], commandValue)
			m.kernel.Commands.With(func(value *[]string) *[]string {
				newValue := append(*value, commandValue)
				return &newValue
			})
		}
		commandValue = ""
		commandHistory_ptr = 0
	case "<Backspace>", "<C-<Backspace>>":
		if len(commandValue) > 0 {
			commandValue = commandValue[:len(commandValue)-1]
		}
	case "<Space>":
		if len(commandValue) > 0 {
			commandValue += " "
		}
	default:
		commandValue += event.ID

	}
	m.CommandEntry.Text = fmt.Sprintf("> %s", commandValue)
}
