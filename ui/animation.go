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

type apngPlayer struct {
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

type AnimationOption struct {
	Looping        bool
	AutoClose      bool
	DisplayTime    time.Duration
	MaxDisplayTime time.Duration
	BGColor        sdl.Color
}

type AnimationReturn struct {
	CompletedNormally bool
	LoopCount         int
	PlayedFrames      int
	LastPressedKey    sdl.Keycode
	LastPressedBtn    uint8
	Cancelled         bool
}

// Animation plays an APNG file. Yeah, that's it.
func Animation(filePath string, options AnimationOption) (AnimationReturn, error) {
	window := internal.GetWindow()
	renderer := window.Renderer

	player, err := newAPNGPlayer(renderer, filePath)
	if err != nil {
		return AnimationReturn{Cancelled: true}, err
	}
	defer player.destroy()

	player.isLooping = true

	screenW, screenH, _ := renderer.GetOutputSize()
	rect := player.getRect()

	x := (screenW - rect.W) / 2

	y := (screenH - rect.H) / 2

	player.setPosition(x, y)

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
						player.isPlaying = !player.isPlaying
					case sdl.K_r:
						player.reset()
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
						player.isPlaying = !player.isPlaying
					case BrickButton_X:
						player.reset()
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

		player.update()

		if player.isPlaying && player.isLooping {
			if options.AutoClose {

				if time.Since(startTime) > options.DisplayTime && options.DisplayTime > 0 {
					running = false
					result.CompletedNormally = true
				}
			}
		}

		if options.MaxDisplayTime > 0 && time.Since(startTime) > options.MaxDisplayTime {
			running = false
			result.CompletedNormally = true
		}

		renderer.SetDrawColor(options.BGColor.R, options.BGColor.G, options.BGColor.B, options.BGColor.A)
		renderer.Clear()

		player.render()

		renderer.Present()
		sdl.Delay(16)
	}

	return result, nil
}

func newAPNGPlayer(renderer *sdl.Renderer, filePath string) (*apngPlayer, error) {
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

		img := image.NewRGBA(image.Rect(0, 0, width, height))

		draw.Draw(img, img.Bounds(), frame.Image, bounds.Min, draw.Over)

		frames[i] = img

		numDenomMs := 1000 * float64(frame.DelayNumerator) / float64(frame.DelayDenominator)
		frameTimes[i] = time.Duration(numDenomMs) * time.Millisecond
	}

	updateTextureFromFrame(texture, frames[0])

	player := &apngPlayer{
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

func (p *apngPlayer) setPosition(x, y int32) {
	p.rect.X = x
	p.rect.Y = y
}

func (p *apngPlayer) getRect() sdl.Rect {
	return p.rect
}

func (p *apngPlayer) setScale(width, height int32) {
	p.rect.W = width
	p.rect.H = height
}

func (p *apngPlayer) reset() {
	p.currentFrame = 0
	p.lastUpdate = time.Now()
	updateTextureFromFrame(p.texture, p.frames[0])
}

func (p *apngPlayer) update() bool {
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

		updateTextureFromFrame(p.texture, p.frames[p.currentFrame])
		p.lastUpdate = now
		frameChanged = true
	}

	return frameChanged
}

func (p *apngPlayer) render() {
	p.renderer.Copy(p.texture, nil, &p.rect)
}

func (p *apngPlayer) destroy() {
	if p.texture != nil {
		p.texture.Destroy()
	}
}

func updateTextureFromFrame(texture *sdl.Texture, img image.Image) error {
	rgba, ok := img.(*image.RGBA)
	if !ok {

		bounds := img.Bounds()
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	}

	pitch := rgba.Stride

	return texture.Update(nil, unsafe.Pointer(&rgba.Pix[0]), pitch)
}
