package scenes

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"nextui-sdl2/models"
	"nextui-sdl2/ui"
)

type MenuScene struct {
	listController *ui.ListController
	renderer       *sdl.Renderer
}

func NewMenuScene(renderer *sdl.Renderer) *MenuScene {
	scene := &MenuScene{
		renderer: renderer,
	}

	menuItems := []models.MenuItem{
		{"Megathread", false, false, nil},
		{"SMB", false, false, nil},
		{"RomM", false, false, nil},
		{"nginx", false, false, nil},
		{"Apache", false, false, nil},
		{"Potato", false, false, nil},
		{"Salad", false, false, nil},
		{"Oh look", false, false, nil},
		{"There is more", false, false, nil},
		{"This scrolls?", false, false, nil},
		{"This scrolls!", false, false, nil},
		{"And its in Go!?", false, false, nil},
		{"Wait does that mean?", false, false, nil},
		{"Can this be a library?", false, false, nil},
		{"Hopefully.", false, false, nil},
		{"Should make things so much smoother.", false, false, nil},
		{"Better UX / UI", false, false, nil},
		{"Better Dev Experience", false, false, nil},
		{"START FROM THE TOP YO!", false, false, nil},
	}

	scene.listController = ui.NewListController("Mortar", menuItems, 20)

	scene.listController.OnSelect = func(index int, item *models.MenuItem) {
		fmt.Printf("YDIJ: %s\n", item.Text)
	}

	scene.listController.MaxVisibleItems = 7

	return scene
}

func (s *MenuScene) Init() error {
	// Any additional initialization (if needed)
	return nil
}

func (s *MenuScene) HandleEvent(event sdl.Event) bool {
	return s.listController.HandleEvent(event)
}

func (s *MenuScene) Update() error {
	// Update logic
	return nil
}

func (s *MenuScene) Render() error {
	s.renderer.SetDrawColor(0, 0, 0, 255)
	s.renderer.Clear()

	s.listController.Render(s.renderer) // This line was missing!
	return nil
}

func (s *MenuScene) Destroy() error {
	// Clean up resources
	return nil
}
