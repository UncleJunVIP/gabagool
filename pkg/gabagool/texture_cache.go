package gabagool

import "github.com/veandco/go-sdl2/sdl"

type textureCache struct {
	textures map[string]*sdl.Texture
}

func newTextureCache() *textureCache {
	return &textureCache{
		textures: make(map[string]*sdl.Texture),
	}
}

func (c *textureCache) get(key string) *sdl.Texture {
	if texture, exists := c.textures[key]; exists {
		return texture
	}
	return nil
}

func (c *textureCache) set(key string, texture *sdl.Texture) {
	c.textures[key] = texture
}

func (c *textureCache) destroy() {
	for _, texture := range c.textures {
		texture.Destroy()
	}
	c.textures = make(map[string]*sdl.Texture)
}
