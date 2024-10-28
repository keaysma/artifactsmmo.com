package runtimes

import (
	"log"
	"os"
	"time"

	"artifactsmmo.com/m/api"
	mainframe "artifactsmmo.com/m/gui"
	gui "artifactsmmo.com/m/gui/backend"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
	ui "github.com/keaysma/termui/v3"
	"github.com/keaysma/termui/v3/widgets"
)

var s = utils.GetSettings()

func UI() {
	err := ui.Init()
	if err != nil {
		log.Fatalf("failed to initialize termui: %s", err)
	}

	defer ui.Close()

	tabs := widgets.NewTabPane("mainframe", "charts", "md")
	tabs.Border = true

	mainframeWidgets := mainframe.Init(s)

	char, err := api.GetCharacterByName(s.Character)
	if err != nil {
		log.Fatalf("failed to get character info for %s: %s", s.Character, err)
		os.Exit(1)
	}
	if char == nil {
		log.Fatalf("char is nil: %s", err)
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

	draw := func() {
		switch tabs.ActiveTabIndex {
		case 0:
			mainframeWidgets.Draw()
		case 1:
		case 2:
		default:
		}

		ui.Render(tabs)
	}

	resize := func(w int, h int) {
		tabHeight := 3
		tabs.SetRect(0, 0, w, tabHeight)

		switch tabs.ActiveTabIndex {
		case 0:
			mainframeWidgets.ResizeWidgets(w, h)
		case 1:
		case 2:
		default:
		}
	}

	w, h := ui.TerminalDimensions()
	resize(w, h)
	draw()

	heavy := 0
	loop := func() {
		switch tabs.ActiveTabIndex {
		case 0:
			mainframeWidgets.Wigets.Loop(heavy == 0)
		case 1:
		case 2:
		default:
		}

		heavy = (heavy + 1) % 6
	}

	uiEvents := ui.PollEvents()
	for {
		select {
		case event := <-uiEvents:
			switch event.Type {
			case ui.ResizeEvent:
				payload := event.Payload.(ui.Resize)
				resize(payload.Width, payload.Height)

			case ui.KeyboardEvent:
				switch event.ID {
				// no-ops
				case "<Escape>":
				case "<C-c>", "<C-v>":
				case "<Left>":
					tabs.ActiveTabIndex = (tabs.ActiveTabIndex - 1 + len(tabs.TabNames)) % len(tabs.TabNames)
				case "<Right>":
					tabs.ActiveTabIndex = (tabs.ActiveTabIndex + 1) % len(tabs.TabNames)
				default:
					mainframeWidgets.HandleKeyboardInput(event)
				}
			}
		default:
		}

		loop()
		draw()

		time.Sleep(50_000_000)

	}
}
