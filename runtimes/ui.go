package runtimes

import (
	"log"
	"os"
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/gui/backend"
	"artifactsmmo.com/m/gui/mainframe"
	"artifactsmmo.com/m/state"
	"artifactsmmo.com/m/types"
	"artifactsmmo.com/m/utils"
	ui "github.com/keaysma/termui/v3"
	"github.com/keaysma/termui/v3/widgets"
)

var s = utils.GetSettings()

type GUIWidget struct {
	Draw                func()
	ResizeWidgets       func(w int, h int)
	Loop                func(heavy bool)
	HandleKeyboardInput func(event ui.Event)
}

func (g *GUIWidget) WidgetList() []ui.Drawable {
	return []ui.Drawable{}
}

var wxs = []GUIWidget{}

func UI() {
	err := ui.Init()
	if err != nil {
		log.Fatalf("failed to initialize termui: %s", err)
	}

	defer ui.Close()

	// conn, err := db.NewDBConnection()
	// if err != nil {
	// 	panic(err)
	// }

	// tabs := widgets.NewTabPane("mainframe", "charts", "md")
	tabs := widgets.NewTabPane("mainframe")
	tabs.Border = true

	mainframeWidgets := mainframe.Init(s)
	wxs = append(wxs, GUIWidget{
		Draw:                mainframeWidgets.Draw,
		ResizeWidgets:       mainframeWidgets.ResizeWidgets,
		Loop:                mainframeWidgets.Loop,
		HandleKeyboardInput: mainframeWidgets.HandleKeyboardInput,
	})

	// chartsWidgets := charts.Init(s, conn)
	// wxs = append(wxs, GUIWidget{
	// 	Draw:                chartsWidgets.Draw,
	// 	ResizeWidgets:       chartsWidgets.ResizeWidgets,
	// 	Loop:                chartsWidgets.Loop,
	// 	HandleKeyboardInput: chartsWidgets.HandleKeyboardInput,
	// })

	char, err := api.GetCharacterByName(s.Character)
	if err != nil {
		log.Fatalf("failed to get character info for %s: %s", s.Character, err)
		os.Exit(1)
	}
	if char == nil {
		log.Fatalf("char is nil: %s", err)
	}
	state.GlobalCharacter.With(func(value *types.Character) *types.Character { return char })

	// If this fails let's just ignore it, not critical
	end, err := time.Parse(time.RFC3339, char.Cooldown_expiration)
	if err == nil {
		state.GlobalCooldown.Set(&state.CooldownData{
			Duration_seconds: char.Cooldown,
			End:              &end,
		})
	}

	go backend.Gameloop()

	draw := func() {
		wxs[tabs.ActiveTabIndex].Draw()
		ui.Render(tabs)
	}

	resize := func(w int, h int) {
		tabHeight := 3
		tabs.SetRect(0, 0, w, tabHeight)
		wxs[tabs.ActiveTabIndex].ResizeWidgets(w, h)
	}

	w, h := ui.TerminalDimensions()
	resize(w, h)
	draw()

	heavy := 0
	loop := func() {
		wxs[tabs.ActiveTabIndex].Loop(heavy == 0)
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
					resize(ui.TerminalDimensions())
				case "<Right>":
					tabs.ActiveTabIndex = (tabs.ActiveTabIndex + 1) % len(tabs.TabNames)
					resize(ui.TerminalDimensions())
				default:
					wxs[tabs.ActiveTabIndex].HandleKeyboardInput(event)
				}
			}
		default:
		}

		loop()
		draw()

		time.Sleep(50_000_000)

	}
}
