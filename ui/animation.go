package ui

import (
	"errors"
	"github.com/UncleJunVIP/gabagool/internal"
	"image"
	"image/draw"
	"os"
	"time"
	"unsafe"

	"github.com/kettek/apng"
	"github.com/veandco/go-sdl2/sdl"
)

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

type AnimationReturn struct {
	CompletedNormally bool
	LoopCount         int
	PlayedFrames      int
	LastPressedKey    sdl.Keycode
	LastPressedBtn    uint8
	Cancelled         bool
}

func NewAPNGPlayer(renderer *sdl.Renderer, filePath string) (*APNGPlayer, error) {
	if renderer == nil {
		return nil, errors.New("renderer cannot be nil")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	animation, err := apng.DecodeAll(file)
	if err != nil {
		return nil, err
	}

	if len(animation.Frames) == 0 {
		return nil, errors.New("no frames found in the APNG file")
	}

	frames := make([]image.Image, len(animation.Frames))
	frameTimes := make([]time.Duration, len(animation.Frames))

	bounds := animation.Frames[0].Image.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STREAMING,
		int32(width),
		int32(height),
	)
	if err != nil {
		return nil, err
	}

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

func (p *APNGPlayer) SetPosition(x, y int32) {
	p.rect.X = x
	p.rect.Y = y
}

func (p *APNGPlayer) GetRect() sdl.Rect {
	return p.rect
}

func (p *APNGPlayer) SetScale(width, height int32) {
	p.rect.W = width
	p.rect.H = height
}

func (p *APNGPlayer) SetLooping(looping bool) {
	p.isLooping = looping
}

func (p *APNGPlayer) IsLooping() bool {
	return p.isLooping
}

func (p *APNGPlayer) Play() {
	p.isPlaying = true
}

func (p *APNGPlayer) Pause() {
	p.isPlaying = false
}

func (p *APNGPlayer) IsPlaying() bool {
	return p.isPlaying
}

func (p *APNGPlayer) Reset() {
	p.currentFrame = 0
	p.lastUpdate = time.Now()
	updateTextureFromFrame(p.renderer, p.texture, p.frames[0])
}

func (p *APNGPlayer) Update() bool {
	if !p.isPlaying {
		return false
	}

	frameChanged := false
	now := time.Now()

	if now.Sub(p.lastUpdate) >= p.frameTimes[p.currentFrame] {
		p.currentFrame++

		if p.currentFrame >= len(p.frames) {
			if p.isLooping {
				p.currentFrame = 0
			} else {
				p.currentFrame = len(p.frames) - 1
				p.isPlaying = false
			}
		}

		updateTextureFromFrame(p.renderer, p.texture, p.frames[p.currentFrame])
		p.lastUpdate = now
		frameChanged = true
	}

	return frameChanged
}

func (p *APNGPlayer) Render() {
	p.renderer.Copy(p.texture, nil, &p.rect)
}

func (p *APNGPlayer) Destroy() {
	if p.texture != nil {
		p.texture.Destroy()
	}
}

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

type animationConfig struct {
	x              int32
	y              int32
	width          int32
	height         int32
	loop           bool
	autoClose      bool
	displayTime    time.Duration
	maxDisplayTime time.Duration
	bgColor        sdl.Color
}

func defaultAnimationConfig() *animationConfig {
	return &animationConfig{
		x:              -1, // Center horizontally
		y:              -1, // Center vertically
		width:          0,  // Original size
		height:         0,  // Original size
		loop:           false,
		autoClose:      true,
		displayTime:    time.Second * 2,                     // Display for 2 seconds after completion
		maxDisplayTime: 0,                                   // No maximum display time
		bgColor:        sdl.Color{R: 0, G: 0, B: 0, A: 255}, // Black background
	}
}

type AnimationOption func(*animationConfig)

func WithPosition(x, y int32) AnimationOption {
	return func(c *animationConfig) {
		c.x = x
		c.y = y
	}
}

func WithCenteredPosition() AnimationOption {
	return func(c *animationConfig) {
		c.x = -1
		c.y = -1
	}
}

func WithScale(width, height int32) AnimationOption {
	return func(c *animationConfig) {
		c.width = width
		c.height = height
	}
}

func WithLooping(loop bool) AnimationOption {
	return func(c *animationConfig) {
		c.loop = loop
	}
}

func WithAutoClose(autoClose bool) AnimationOption {
	return func(c *animationConfig) {
		c.autoClose = autoClose
	}
}

func WithDisplayTime(duration time.Duration) AnimationOption {
	return func(c *animationConfig) {
		c.displayTime = duration
	}
}

func WithMaxDisplayTime(duration time.Duration) AnimationOption {
	return func(c *animationConfig) {
		c.maxDisplayTime = duration
	}
}

func WithBackgroundColor(color sdl.Color) AnimationOption {
	return func(c *animationConfig) {
		c.bgColor = color
	}
}

func NewBlockingAnimation(filePath string, options ...AnimationOption) (AnimationReturn, error) {
	window := internal.GetWindow()
	renderer := window.Renderer

	player, err := NewAPNGPlayer(renderer, filePath)
	if err != nil {
		return AnimationReturn{Cancelled: true}, err
	}
	defer player.Destroy()

	config := defaultAnimationConfig()
	for _, option := range options {
		option(config)
	}

	player.SetLooping(config.loop)

	if config.x == -1 || config.y == -1 {
		screenW, screenH, _ := renderer.GetOutputSize()
		rect := player.GetRect()

		x := config.x
		if x == -1 {
			x = (screenW - rect.W) / 2
		}

		y := config.y
		if y == -1 {
			y = (screenH - rect.H) / 2
		}

		player.SetPosition(x, y)
	} else {
		player.SetPosition(config.x, config.y)
	}

	if config.width > 0 && config.height > 0 {
		player.SetScale(config.width, config.height)
	}

	result := AnimationReturn{
		CompletedNormally: false,
		LoopCount:         0,
		PlayedFrames:      0,
		LastPressedKey:    0,
		LastPressedBtn:    0,
		Cancelled:         false,
	}

	running := true
	prevFrame := player.currentFrame
	startTime := time.Now()

	loopCount := 0
	lastFrameIndex := len(player.frames) - 1

	helpLines := []string{
		"Animation Controls",
		"Space/A: Pause/Resume",
		"Esc/B: Exit animation",
		"R/X: Reset animation",
		"H/Menu: Show/hide help",
	}
	showingHelp := false

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				result.Cancelled = true

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					result.LastPressedKey = e.Keysym.Sym

					if showingHelp {
						showingHelp = false
						continue
					}

					switch e.Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false
						result.Cancelled = true
					case sdl.K_SPACE:
						if player.IsPlaying() {
							player.Pause()
						} else {
							player.Play()
						}
					case sdl.K_r:
						player.Reset()
						result.PlayedFrames = 0
						loopCount = 0
					case sdl.K_h:
						showingHelp = !showingHelp
					}
				}

			case *sdl.ControllerButtonEvent:
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					result.LastPressedBtn = e.Button

					if showingHelp {
						showingHelp = false
						continue
					}

					switch e.Button {
					case BrickButton_B:
						running = false
						result.Cancelled = true
					case BrickButton_A:
						if player.IsPlaying() {
							player.Pause()
						} else {
							player.Play()
						}
					case BrickButton_X:
						player.Reset()
						result.PlayedFrames = 0
						loopCount = 0
					case BrickButton_MENU:
						showingHelp = !showingHelp
					}
				}
			}
		}

		if player.currentFrame != prevFrame {
			result.PlayedFrames++

			if prevFrame == lastFrameIndex && player.currentFrame == 0 {
				loopCount++
				result.LoopCount = loopCount
			}

			prevFrame = player.currentFrame
		}

		player.Update()

		if !player.IsPlaying() && !player.IsLooping() {
			if config.autoClose {
				// Wait for the configured delay before auto-closing
				if time.Since(startTime) > config.displayTime && config.displayTime > 0 {
					running = false
					result.CompletedNormally = true
				}
			}
		}

		if config.maxDisplayTime > 0 && time.Since(startTime) > config.maxDisplayTime {
			running = false
			result.CompletedNormally = true
		}

		renderer.SetDrawColor(config.bgColor.R, config.bgColor.G, config.bgColor.B, config.bgColor.A)
		renderer.Clear()

		player.Render()

		if showingHelp {
			renderHelpOverlay(renderer, helpLines)
		}

		renderer.Present()
		sdl.Delay(16)
	}

	return result, nil
}

func renderHelpOverlay(renderer *sdl.Renderer, helpLines []string) {
	screenW, screenH, _ := renderer.GetOutputSize()
	overlay := sdl.Rect{X: 0, Y: 0, W: screenW, H: screenH}
	renderer.SetDrawColor(0, 0, 0, 200)
	renderer.FillRect(&overlay)

	font := internal.GetTitleFont()
	titleSurface, err := font.RenderUTF8Solid("Animation Help", sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err == nil {
		texture, err := renderer.CreateTextureFromSurface(titleSurface)
		if err == nil {
			rect := &sdl.Rect{
				X: (screenW - titleSurface.W) / 2,
				Y: screenH/6 - titleSurface.H/2,
				W: titleSurface.W,
				H: titleSurface.H,
			}
			renderer.Copy(texture, nil, rect)
			texture.Destroy()
		}
		titleSurface.Free()
	}

	normalFont := internal.GetFont()
	lineHeight := int32(30)
	startY := screenH/6 + lineHeight*2

	for i, line := range helpLines {
		lineSurface, err := normalFont.RenderUTF8Solid(line, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if err == nil {
			texture, err := renderer.CreateTextureFromSurface(lineSurface)
			if err == nil {
				rect := &sdl.Rect{
					X: (screenW - lineSurface.W) / 2,
					Y: startY + int32(i)*lineHeight,
					W: lineSurface.W,
					H: lineSurface.H,
				}
				renderer.Copy(texture, nil, rect)
				texture.Destroy()
			}
			lineSurface.Free()
		}
	}

	dismissText := "Press any key to dismiss"
	dismissSurface, err := normalFont.RenderUTF8Solid(dismissText, sdl.Color{R: 180, G: 180, B: 180, A: 255})
	if err == nil {
		texture, err := renderer.CreateTextureFromSurface(dismissSurface)
		if err == nil {
			rect := &sdl.Rect{
				X: (screenW - dismissSurface.W) / 2,
				Y: screenH - dismissSurface.H - 40,
				W: dismissSurface.W,
				H: dismissSurface.H,
			}
			renderer.Copy(texture, nil, rect)
			texture.Destroy()
		}
		dismissSurface.Free()
	}
}
