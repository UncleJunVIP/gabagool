package internal

import (
	"github.com/veandco/go-sdl2/sdl"
	"nextui-sdl2/internal/types"
)

type SceneManager struct {
	scenes         map[string]types.Scene
	currentScene   types.Scene
	currentSceneID string
	window         *Window
}

var sceneManagerInstance *SceneManager

func NewSceneManager() *SceneManager {
	if sceneManagerInstance != nil {
		return sceneManagerInstance
	}

	sceneManagerInstance = &SceneManager{
		scenes:       make(map[string]types.Scene),
		currentScene: nil,
		window:       GetWindow(),
	}
	return sceneManagerInstance
}

func GetSceneManager() *SceneManager {
	return sceneManagerInstance
}

func (sm *SceneManager) AddScene(id string, scene types.Scene) {
	sm.scenes[id] = scene
}

func (sm *SceneManager) GetScene(id string) types.Scene {
	return sm.scenes[id]
}

func (sm *SceneManager) SwitchTo(id string) error {
	scene, exists := sm.scenes[id]
	if !exists {
		return nil
	}

	if sm.currentScene != nil {
		err := sm.currentScene.Deactivate()
		if err != nil {
			return err
		}
	}

	sm.currentScene = scene
	sm.currentSceneID = id

	return scene.Activate()
}

func (sm *SceneManager) HandleEvent(event sdl.Event) bool {
	if sm.currentScene != nil {
		return sm.currentScene.HandleEvent(event)
	}
	return false
}

func (sm *SceneManager) Update() error {
	if sm.currentScene != nil {
		return sm.currentScene.Update()
	}
	return nil
}

func (sm *SceneManager) Render() error {
	if sm.currentScene != nil {
		return sm.currentScene.Render()
	}
	return nil
}

func (sm *SceneManager) GetCurrentSceneID() string {
	return sm.currentSceneID
}

func (sm *SceneManager) DestroyScene(name string) error {
	scene, exists := sm.scenes[name]
	if !exists {
		return nil
	}

	res := scene.Destroy()
	if res != nil {
		return res
	}

	delete(sm.scenes, name)
	return nil
}
