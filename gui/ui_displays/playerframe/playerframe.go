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

var s = utils.GetSettings()
var TAB_HEIGHT = s.TabHeight

var commandValue = ""

// var logLines = []string{}
var commandHistory = []string{}
var commandHistory_ptr = 0

type Mainframe struct {
	// TODO: Do this... differently?
	loglines []string

	kernel                    *game.Kernel
	Logs                      *widgets.Paragraph
	CommandList               *widgets.Paragraph
	OrderReferenceList        *widgets.Paragraph
	EquipmentDisplay          *widgets.Table
	InventoryDisplay          *widgets.List
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
	GaugeSkillAlchemy         *widgets.Gauge
	BankItemList              *widgets.List
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

	equipmentDisplay := widgets.NewTable()
	equipmentDisplay.Title = "Equipment"
	equipmentDisplay.Rows = [][]string{
		{"", ""},
		{"", ""},
	}

	inventoryDisplay := widgets.NewList()
	inventoryDisplay.Title = "Inventory"
	inventoryDisplay.Rows = []string{
		"",
	}

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
	gauge_skill_alchemy := widgets.NewGauge()

	bankItemsList := widgets.NewList()

	mainframeWigets := Mainframe{
		kernel:                    kernel,
		Logs:                      logs,
		CommandList:               commandList,
		OrderReferenceList:        orderReferenceList,
		EquipmentDisplay:          equipmentDisplay,
		InventoryDisplay:          inventoryDisplay,
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
		GaugeSkillAlchemy:         gauge_skill_alchemy,
		BankItemList:              bankItemsList,
	}

	return &mainframeWigets
}

func (m *Mainframe) Draw() {
	ui.Render(m.Logs, m.CommandList, m.CharacterDisplay, m.CooldownGauge, m.CommandEntry, m.GaugeSkillMining, m.GaugeSkillWoodcutting, m.GaugeSkillFishing, m.GaugeSkillWeaponcrafting, m.GaugeSkillGearcrafting, m.GaugeSkillJewelrycrafting, m.GaugeSkillCooking, m.GaugeSkillAlchemy, m.EquipmentDisplay)

	if m.OrderReferenceList.Text != "" {
		ui.Render(m.OrderReferenceList)
	} else {
		ui.Render(m.InventoryDisplay)
	}

	if m.kernel.BankItemListShown {
		ui.Render(m.BankItemList)
	}
}

func (m *Mainframe) ResizeWidgets(w int, h int) {
	mid_w := w - 72      // w/2
	last_w := mid_w + 42 // w-(w/4) aka last_w

	first_h := 14
	mid_h := h - 21 - 16

	m.Logs.SetRect(0, TAB_HEIGHT, mid_w, h-3)

	// swap commands and equipment (it does look weird though...)
	// m.CommandList.SetRect(mid_w, TAB_HEIGHT, last_w-1, mid_h)
	m.CommandList.SetRect(last_w, first_h+24, w, h-6)
	m.OrderReferenceList.SetRect(mid_w, h-18, last_w-1, h-6)
	m.InventoryDisplay.SetRect(mid_w, mid_h, last_w-1, h-6)

	m.CharacterDisplay.SetRect(last_w, TAB_HEIGHT, w, first_h)
	m.CooldownGauge.SetRect(mid_w, h-6, w, h-3)

	m.GaugeSkillMining.SetRect(last_w, first_h, w, first_h+3)
	m.GaugeSkillWoodcutting.SetRect(last_w, first_h+3, w, first_h+6)
	m.GaugeSkillFishing.SetRect(last_w, first_h+6, w, first_h+9)
	m.GaugeSkillWeaponcrafting.SetRect(last_w, first_h+9, w, first_h+12)
	m.GaugeSkillGearcrafting.SetRect(last_w, first_h+12, w, first_h+15)
	m.GaugeSkillJewelrycrafting.SetRect(last_w, first_h+15, w, first_h+18)
	m.GaugeSkillCooking.SetRect(last_w, first_h+18, w, first_h+21)
	m.GaugeSkillAlchemy.SetRect(last_w, first_h+21, w, first_h+24)

	// m.EquipmentDisplay.SetRect(last_w, first_h+24, w, h-6)
	m.EquipmentDisplay.SetRect(mid_w, TAB_HEIGHT, last_w-1, mid_h)

	m.BankItemList.SetRect(mid_w-36, TAB_HEIGHT+1, mid_w-1, h-4)

	m.CommandEntry.SetRect(0, h-3, w, h)
}

// For rendering-related tasks that should still happen
// when the current mainframe is not visible
func (m *Mainframe) BackgroundLoop() {
	select {
	case line := <-m.kernel.LogsChannel:
		m.loglines = append(m.loglines, line)
	case line := <-utils.LogsChannel:
		m.loglines = append(m.loglines, line)
	default:
	}
	h := m.Logs.Inner.Dy()
	if len(m.loglines) > h {
		m.loglines = m.loglines[max(0, len(m.loglines)-h):]
	}
	m.Logs.Text = strings.Join(m.loglines, "\n")
}

func (m *Mainframe) Loop(heavy bool) {
	m.CommandList.Text = strings.Join(m.kernel.Commands.ShallowCopy(), "\n")

	generator_name := m.kernel.CurrentGeneratorName.ShallowCopy()
	if generator_name != "" {
		m.CommandList.Title = fmt.Sprintf("Commands (gen: %s)", generator_name)
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

		{
			character := m.kernel.CharacterState.Ref()

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

			m.GaugeSkillAlchemy.Title = fmt.Sprintf("Alchemy: %d", character.Alchemy_level)
			m.GaugeSkillAlchemy.Percent = int((float64(character.Alchemy_xp) / float64(character.Alchemy_max_xp)) * 100)

			newList := []string{}
			for _, item := range character.Inventory {
				entry := fmt.Sprintf("(%d) %s", item.Quantity, item.Code)
				newList = append(newList, entry)
			}
			m.InventoryDisplay.Rows = newList

			newEquipmentTable := [][]string{
				{"weapon", character.Weapon_slot},
				{"shield", character.Shield_slot},
				{"helmet", character.Helmet_slot},
				{"body_armor", character.Body_armor_slot},
				{"legs_armor", character.Leg_armor_slot},
				{"amulet", character.Amulet_slot},
				{"ring1", character.Ring1_slot},
				{"ring2", character.Ring2_slot},
				{"utility1", character.Utility1_slot},
				{"utility2", character.Utility2_slot},
				{"artifact1", character.Artifact1_slot},
			}
			m.EquipmentDisplay.Rows = newEquipmentTable

			m.kernel.CharacterState.Unlock()
		}

		m.OrderReferenceList.Text = ""
		{
			ordersList := state.OrderIdsReference.Ref()
			for i, id := range *ordersList {
				m.OrderReferenceList.Text += fmt.Sprintf("%d: %s\n", i, id)
			}

			state.OrderIdsReference.Unlock()
		}

		if m.kernel.BankItemListShown {
			newBankItems := []string{}
			{
				bankItems := state.GlobalState.BankState.Ref()

				for _, item := range *bankItems {
					if m.kernel.BankItemListFilter == nil || strings.Contains(item.Code, *m.kernel.BankItemListFilter) {
						newBankItems = append(newBankItems, fmt.Sprintf("(%6d) %s", item.Quantity, item.Code))
					}
				}

				state.GlobalState.BankState.Unlock()
			}
			m.BankItemList.Rows = newBankItems
		}
	}
}

var PRIORITY_COMMANDS = []string{"o", "myo", "simulate-fight", "list-bank", "hide-bank"}

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
			m.kernel.Log("help message")
		} else if commandValue == "clear" {
			m.loglines = []string{}
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
