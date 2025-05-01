package ui

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
)

type Scene interface {
	Init() error                      // Initialize scene resources
	HandleEvent(event sdl.Event) bool // Process events, return true if handled
	Update() error                    // Update scene logic
	Render() error                    // Render the scene
	Destroy() error                   // Clean up resources
}

type SceneManager struct {
	scenes       map[string]Scene
	currentScene string
	window       *Window
}

func NewSceneManager(window *Window) *SceneManager {
	return &SceneManager{
		scenes: make(map[string]Scene),
		window: window,
	}
}

func (sm *SceneManager) AddScene(name string, scene Scene) {
	sm.scenes[name] = scene
}

func (sm *SceneManager) SwitchTo(name string) error {
	if _, exists := sm.scenes[name]; !exists {
		return fmt.Errorf("scene %s does not exist", name)
	}

	if sm.currentScene != "" {
		if err := sm.scenes[sm.currentScene].Destroy(); err != nil {
			return err
		}
	}

	sm.currentScene = name
	return sm.scenes[name].Init()
}

func (sm *SceneManager) HandleEvent(event sdl.Event) bool {
	if sm.currentScene == "" {
		return false
	}
	return sm.scenes[sm.currentScene].HandleEvent(event)
}

func (sm *SceneManager) Update() error {
	if sm.currentScene == "" {
		return nil
	}
	return sm.scenes[sm.currentScene].Update()
}

func (sm *SceneManager) Render() error {
	if sm.currentScene == "" {
		return nil
	}
	return sm.scenes[sm.currentScene].Render()
}

func (sm *SceneManager) CurrentSceneName() string {
	return sm.currentScene
}
