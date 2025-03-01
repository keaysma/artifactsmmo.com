package runtimes

import (
	"log"
	"time"

	"artifactsmmo.com/m/api"
	"artifactsmmo.com/m/game"
	"artifactsmmo.com/m/gui/backend"
	"artifactsmmo.com/m/gui/ui_displays/playerframe"
	"artifactsmmo.com/m/utils"
	ui "github.com/keaysma/termui/v3"
	"github.com/keaysma/termui/v3/widgets"
)

var s = utils.GetSettings()
var kernels = map[string]*game.Kernel{}

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
	var err error

	characters, err := api.GetAllCharacters()
	if err != nil {
		log.Fatalf("failed to get characters: %s", err)
	}

	err = ui.Init()
	if err != nil {
		log.Fatalf("failed to initialize termui: %s", err)
	}

	defer ui.Close()

	// conn, err := db.NewDBConnection()
	// if err != nil {
	// 	panic(err)
	// }

	characterNames := []string{}
	for _, character := range *characters {
		characterNames = append(characterNames, character.Name)
		kernels[character.Name] = backend.NewKernel(character)
	}

	tabs := widgets.NewTabPane(characterNames...)
	tabs.Border = true

	for _, characterName := range characterNames {
		mainframeWidgets := playerframe.Init(s, kernels[characterName])
		wxs = append(wxs, GUIWidget{
			Draw:                mainframeWidgets.Draw,
			ResizeWidgets:       mainframeWidgets.ResizeWidgets,
			Loop:                mainframeWidgets.Loop,
			HandleKeyboardInput: mainframeWidgets.HandleKeyboardInput,
		})
	}

	// go backend.Gameloop()
	// go backend.PriorityLoop(backend.PriorityCommands)
	for _, characterName := range characterNames {
		kernel := kernels[characterName]
		go backend.Gameloop(kernel)
		go backend.PriorityLoop(kernel)
	}

	heavy := 0
	draw := func() {
		ui.Render(tabs)
		wxs[tabs.ActiveTabIndex].Draw()
		wxs[tabs.ActiveTabIndex].Loop(heavy == 0)
	}

	resize := func(w int, h int) {
		tabHeight := 3
		tabs.SetRect(0, 0, w, tabHeight)
		wxs[tabs.ActiveTabIndex].ResizeWidgets(w, h)
	}

	w, h := ui.TerminalDimensions()
	resize(w, h)
	draw()

	uiEvents := ui.PollEvents()
	for {
		select {
		case event := <-uiEvents:
			switch event.Type {
			case ui.ResizeEvent:
				resize(ui.TerminalDimensions())
				draw()
			case ui.KeyboardEvent:
				switch event.ID {
				// no-ops
				case "<Escape>":
				case "<C-c>", "<C-v>":
				case "<Left>":
					heavy = 0
					tabs.ActiveTabIndex = (tabs.ActiveTabIndex - 1 + len(tabs.TabNames)) % len(tabs.TabNames)
					resize(ui.TerminalDimensions())
					draw()
				case "<Right>":
					heavy = 0
					tabs.ActiveTabIndex = (tabs.ActiveTabIndex + 1) % len(tabs.TabNames)
					resize(ui.TerminalDimensions())
					draw()
				default:
					wxs[tabs.ActiveTabIndex].HandleKeyboardInput(event)
				}
			}
		default:
		}

		heavy = (heavy + 1) % 6
		draw()

		time.Sleep(50_000_000)

	}
}
