package internal

import "github.com/veandco/go-sdl2/sdl"

type TextureCache struct {
	textures map[string]*sdl.Texture
}

func NewTextureCache() *TextureCache {
	return &TextureCache{
		textures: make(map[string]*sdl.Texture),
	}
}

func (c *TextureCache) Get(key string) *sdl.Texture {
	if texture, exists := c.textures[key]; exists {
		return texture
	}
	return nil
}

func (c *TextureCache) Set(key string, texture *sdl.Texture) {
	c.textures[key] = texture
}

func (c *TextureCache) Destroy() {
	for _, texture := range c.textures {
		texture.Destroy()
	}
	c.textures = make(map[string]*sdl.Texture)
}
