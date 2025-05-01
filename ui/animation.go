package ui

import (
	"errors"
	"image"
	"image/draw"
	"os"
	"time"
	"unsafe"

	"github.com/kettek/apng"
	"github.com/veandco/go-sdl2/sdl"
)

// APNGPlayer is a reusable component for displaying APNG animations with SDL2
type APNGPlayer struct {
	renderer     *sdl.Renderer
	texture      *sdl.Texture
	animation    apng.APNG
	frames       []image.Image
	frameTimes   []time.Duration
	currentFrame int
	lastUpdate   time.Time
	rect         sdl.Rect
	isPlaying    bool
	isLooping    bool
}

// NewAPNGPlayer creates a new APNG animation player
func NewAPNGPlayer(renderer *sdl.Renderer, filePath string) (*APNGPlayer, error) {
	if renderer == nil {
		return nil, errors.New("renderer cannot be nil")
	}

	// Open the APNG file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode the APNG file
	animation, err := apng.DecodeAll(file)
	if err != nil {
		return nil, err
	}

	if len(animation.Frames) == 0 {
		return nil, errors.New("no frames found in the APNG file")
	}

	// Extract all frames and calculate frame durations
	frames := make([]image.Image, len(animation.Frames))
	frameTimes := make([]time.Duration, len(animation.Frames))

	// Get the dimensions from the first frame
	bounds := animation.Frames[0].Image.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Create a texture for rendering
	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING,
		int32(width),
		int32(height),
	)
	if err != nil {
		return nil, err
	}

	// Extract frames and durations
	for i, frame := range animation.Frames {
		// For each frame, create a new RGBA image with proper dimensions
		img := image.NewRGBA(image.Rect(0, 0, width, height))

		// Draw the frame onto the RGBA image
		draw.Draw(img, img.Bounds(), frame.Image, bounds.Min, draw.Over)

		frames[i] = img

		// Calculate frame duration in milliseconds
		numDenomMs := 1000 * float64(frame.DelayNumerator) / float64(frame.DelayDenominator)
		frameTimes[i] = time.Duration(numDenomMs) * time.Millisecond
	}

	// Update texture with the first frame
	updateTextureFromFrame(renderer, texture, frames[0])

	player := &APNGPlayer{
		renderer:     renderer,
		texture:      texture,
		animation:    animation,
		frames:       frames,
		frameTimes:   frameTimes,
		currentFrame: 0,
		lastUpdate:   time.Now(),
		rect:         sdl.Rect{X: 0, Y: 0, W: int32(width), H: int32(height)},
		isPlaying:    true,
		isLooping:    true,
	}

	return player, nil
}

// SetPosition sets the position of the animation
func (p *APNGPlayer) SetPosition(x, y int32) {
	p.rect.X = x
	p.rect.Y = y
}

// GetRect returns the current SDL rectangle
func (p *APNGPlayer) GetRect() sdl.Rect {
	return p.rect
}

// SetScale sets the scale of the animation
func (p *APNGPlayer) SetScale(width, height int32) {
	p.rect.W = width
	p.rect.H = height
}

// SetLooping sets whether the animation should loop
func (p *APNGPlayer) SetLooping(looping bool) {
	p.isLooping = looping
}

// IsLooping returns whether the animation is set to loop
func (p *APNGPlayer) IsLooping() bool {
	return p.isLooping
}

// Play starts or resumes the animation
func (p *APNGPlayer) Play() {
	p.isPlaying = true
}

// Pause pauses the animation
func (p *APNGPlayer) Pause() {
	p.isPlaying = false
}

// IsPlaying returns whether the animation is currently playing
func (p *APNGPlayer) IsPlaying() bool {
	return p.isPlaying
}

// Reset resets the animation to the first frame
func (p *APNGPlayer) Reset() {
	p.currentFrame = 0
	p.lastUpdate = time.Now()
	updateTextureFromFrame(p.renderer, p.texture, p.frames[0])
}

// Update updates the animation state
func (p *APNGPlayer) Update() bool {
	if !p.isPlaying {
		return false
	}

	frameChanged := false
	now := time.Now()

	// Check if it's time to move to the next frame
	if now.Sub(p.lastUpdate) >= p.frameTimes[p.currentFrame] {
		// Go to next frame
		p.currentFrame++

		// Check if we've reached the end of animation
		if p.currentFrame >= len(p.frames) {
			if p.isLooping {
				p.currentFrame = 0
			} else {
				p.currentFrame = len(p.frames) - 1
				p.isPlaying = false
			}
		}

		// Update texture with new frame
		updateTextureFromFrame(p.renderer, p.texture, p.frames[p.currentFrame])
		p.lastUpdate = now
		frameChanged = true
	}

	return frameChanged
}

// Render draws the current frame of the animation
func (p *APNGPlayer) Render() {
	p.renderer.Copy(p.texture, nil, &p.rect)
}

// Destroy frees all resources used by the player
func (p *APNGPlayer) Destroy() {
	if p.texture != nil {
		p.texture.Destroy()
	}
}

// updateTextureFromFrame updates the SDL texture with the image data
func updateTextureFromFrame(renderer *sdl.Renderer, texture *sdl.Texture, img image.Image) error {
	rgba, ok := img.(*image.RGBA)
	if !ok {
		// Convert to RGBA if it's not already
		bounds := img.Bounds()
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	}

	pitch := rgba.Stride

	return texture.Update(nil, unsafe.Pointer(&rgba.Pix[0]), pitch)
}
