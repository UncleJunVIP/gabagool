package scenes

import (
	"github.com/veandco/go-sdl2/sdl"
	"nextui-sdl2/ui"
	// Include other necessary imports
)

type APNGScene struct {
	renderer      *sdl.Renderer
	apngPlayer    *ui.APNGPlayer // Your APNG player implementation
	isInitialized bool
}

func NewAPNGScene(renderer *sdl.Renderer) *APNGScene {
	return &APNGScene{
		renderer:      renderer,
		isInitialized: false,
	}
}

func (s *APNGScene) Init() error {
	s.apngPlayer, _ = ui.NewAPNGPlayer(s.renderer, "mortar.apng")

	s.isInitialized = true
	return nil
}

func (s *APNGScene) HandleEvent(event sdl.Event) bool {
	return false
}

func (s *APNGScene) Update() error {
	if !s.isInitialized {
		return nil
	}
	s.apngPlayer.Update()
	return nil
}

func (s *APNGScene) Render() error {
	if !s.isInitialized {
		return nil
	}

	s.renderer.SetDrawColor(0, 0, 0, 255)
	s.renderer.Clear()

	// Render the APNG animation
	s.apngPlayer.Render()

	return nil
}

func (s *APNGScene) Destroy() error {
	if !s.isInitialized {
		return nil
	}

	// Properly clean up all resources
	if s.apngPlayer != nil {
		err := s.apngPlayer.Destroy
		if err != nil {
			return nil
		}
		s.apngPlayer = nil
	}

	s.isInitialized = false
	return nil
}
