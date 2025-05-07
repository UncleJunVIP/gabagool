package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/internal"
	"math"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type ConfirmationScreen struct {
	renderer      *sdl.Renderer
	font          *ttf.Font
	message       string
	yesText       string
	noText        string
	visible       bool
	result        bool
	resultSet     bool
	yesSelected   bool
	animationTick float64
	startTime     time.Time
	lastDirection int
	debounceTime  time.Duration
	lastInput     time.Time
}

const (
	DirectionNone = iota
	DirectionLeft
	DirectionRight
)

func NewConfirmationScreen(message string) *ConfirmationScreen {
	window := internal.GetWindow()
	renderer := window.Renderer
	font := internal.GetFont()

	return &ConfirmationScreen{
		renderer:      renderer,
		font:          font,
		message:       message,
		yesText:       "Yes",
		noText:        "No",
		visible:       false,
		result:        false,
		resultSet:     false,
		yesSelected:   false,
		startTime:     time.Now(),
		lastDirection: DirectionNone,
		debounceTime:  time.Millisecond * 200,
		lastInput:     time.Now(),
	}
}

func (c *ConfirmationScreen) Show() {
	c.visible = true
	c.resultSet = false
	c.startTime = time.Now()
}

func (c *ConfirmationScreen) Hide() {
	c.visible = false
}

func (c *ConfirmationScreen) IsVisible() bool {
	return c.visible
}

func (c *ConfirmationScreen) SetMessage(message string) {
	c.message = message
}

func (c *ConfirmationScreen) SetOptions(yesText, noText string) {
	c.yesText = yesText
	c.noText = noText
}

func (c *ConfirmationScreen) GetResult() (bool, bool) {
	return c.result, c.resultSet
}

func (c *ConfirmationScreen) Update(deltaTime float64) {
	if !c.visible {
		return
	}

	c.animationTick += deltaTime * 2
	if c.animationTick > 2*3.14159 {
		c.animationTick = 0
	}
}

func (c *ConfirmationScreen) HandleEvent(event sdl.Event) bool {
	if !c.visible {
		return false
	}

	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		if e.Type == sdl.KEYDOWN {
			switch e.Keysym.Sym {
			case sdl.K_LEFT:
				return c.handleDirection(DirectionLeft)
			case sdl.K_RIGHT:
				return c.handleDirection(DirectionRight)
			case sdl.K_RETURN, sdl.K_SPACE:
				c.result = c.yesSelected
				c.resultSet = true
				c.visible = false
				return true
			case sdl.K_ESCAPE:
				c.result = false
				c.resultSet = true
				c.visible = false
				return true
			}
		}
	case *sdl.MouseButtonEvent:
		if e.Type == sdl.MOUSEBUTTONDOWN && e.Button == sdl.BUTTON_LEFT {
			return c.handleMouseClick(e.X, e.Y)
		}
	case *sdl.MouseMotionEvent:
		return c.handleMouseMotion(e.X, e.Y)
	case *sdl.ControllerButtonEvent:
		if e.Type == sdl.CONTROLLERBUTTONDOWN {
			return c.handleControllerButton(e.Button)
		}
	case *sdl.ControllerAxisEvent:
		return c.handleControllerAxis(e.Axis, e.Value)
	}

	return false
}

func (c *ConfirmationScreen) handleDirection(direction int) bool {
	if time.Since(c.lastInput) < c.debounceTime {
		return false
	}

	c.lastInput = time.Now()

	if direction == DirectionLeft || direction == DirectionRight {
		if direction != c.lastDirection {
			c.yesSelected = !c.yesSelected
			c.lastDirection = direction
			return true
		}
	}

	return false
}

func (c *ConfirmationScreen) handleControllerButton(button uint8) bool {
	switch button {
	case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
		return c.handleDirection(DirectionLeft)
	case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
		return c.handleDirection(DirectionRight)
	case sdl.CONTROLLER_BUTTON_A:
		c.result = c.yesSelected
		c.resultSet = true
		c.visible = false
		return true
	}

	return false
}

func (c *ConfirmationScreen) handleControllerAxis(axis uint8, value int16) bool {
	const deadZone = 8000 // Adjust as needed

	if axis == sdl.CONTROLLER_AXIS_LEFTX {
		if value < -deadZone {
			return c.handleDirection(DirectionLeft)
		} else if value > deadZone {
			return c.handleDirection(DirectionRight)
		} else if c.lastDirection != DirectionNone {
			c.lastDirection = DirectionNone
		}
	}

	return false
}

func (c *ConfirmationScreen) handleMouseClick(x, y int32) bool {
	w, h, err := c.renderer.GetOutputSize()
	if err != nil {
		return false
	}

	buttonWidth := int32(140)
	buttonHeight := int32(50)
	buttonSpacing := int32(30)
	buttonsWidth := 2*buttonWidth + buttonSpacing

	centerX := w / 2
	centerY := h / 2

	noX := centerX - buttonsWidth/2
	yesX := noX + buttonWidth + buttonSpacing
	buttonsY := centerY + int32(20)

	if x >= noX && x <= noX+buttonWidth && y >= buttonsY && y <= buttonsY+buttonHeight {
		c.result = false
		c.resultSet = true
		c.visible = false
		return true
	}

	if x >= yesX && x <= yesX+buttonWidth && y >= buttonsY && y <= buttonsY+buttonHeight {
		c.result = true
		c.resultSet = true
		c.visible = false
		return true
	}

	return false
}

func (c *ConfirmationScreen) handleMouseMotion(x, y int32) bool {
	w, h, err := c.renderer.GetOutputSize()
	if err != nil {
		return false
	}

	buttonWidth := int32(140)
	buttonHeight := int32(50)
	buttonSpacing := int32(30)
	buttonsWidth := 2*buttonWidth + buttonSpacing

	centerX := w / 2
	centerY := h / 2

	noX := centerX - buttonsWidth/2
	yesX := noX + buttonWidth + buttonSpacing
	buttonsY := centerY + int32(20)

	if x >= yesX && x <= yesX+buttonWidth && y >= buttonsY && y <= buttonsY+buttonHeight {
		if !c.yesSelected {
			c.yesSelected = true
			return true
		}
	} else if x >= noX && x <= noX+buttonWidth && y >= buttonsY && y <= buttonsY+buttonHeight {
		if c.yesSelected {
			c.yesSelected = false
			return true
		}
	}

	return false
}

func (c *ConfirmationScreen) Render() {
	if !c.visible {
		return
	}

	w, h, err := c.renderer.GetOutputSize()
	if err != nil {
		fmt.Println("Error getting renderer output size:", err)
		return
	}

	c.renderer.SetDrawColor(0, 0, 0, 200)
	c.renderer.FillRect(&sdl.Rect{X: 0, Y: 0, W: w, H: h})

	centerX := w / 2
	centerY := h / 2

	surface, err := c.font.RenderUTF8Blended(c.message, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		fmt.Println("Error rendering text:", err)
		return
	}
	defer surface.Free()

	texture, err := c.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		fmt.Println("Error creating texture:", err)
		return
	}
	defer texture.Destroy()

	_, _, textWidth, textHeight, err := texture.Query()
	if err != nil {
		fmt.Println("Error querying texture:", err)
		return
	}

	textX := centerX - textWidth/2
	textY := centerY - textHeight - int32(30)
	c.renderer.Copy(texture, nil, &sdl.Rect{X: textX, Y: textY, W: textWidth, H: textHeight})

	buttonWidth := int32(140)
	buttonHeight := int32(50)
	buttonSpacing := int32(30)
	buttonsWidth := 2*buttonWidth + buttonSpacing

	noX := centerX - buttonsWidth/2
	yesX := noX + buttonWidth + buttonSpacing
	buttonsY := centerY + int32(20)

	c.renderButton(noX, buttonsY, buttonWidth, buttonHeight, c.noText, !c.yesSelected)

	c.renderButton(yesX, buttonsY, buttonWidth, buttonHeight, c.yesText, c.yesSelected)

}

func (c *ConfirmationScreen) renderButton(x, y, width, height int32, text string, selected bool) {
	var buttonY = y
	if selected {
		offset := int32(math.Sin(c.animationTick) * 3)
		buttonY = y + offset
	}

	if selected {
		c.renderer.SetDrawColor(100, 100, 200, 255)
	} else {
		c.renderer.SetDrawColor(70, 70, 70, 255)
	}
	c.renderer.FillRect(&sdl.Rect{X: x, Y: buttonY, W: width, H: height})

	c.renderer.SetDrawColor(200, 200, 200, 255)
	c.renderer.DrawRect(&sdl.Rect{X: x, Y: buttonY, W: width, H: height})

	textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	surface, err := c.font.RenderUTF8Blended(text, textColor)
	if err != nil {
		fmt.Println("Error rendering button text:", err)
		return
	}
	defer surface.Free()

	texture, err := c.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		fmt.Println("Error creating button texture:", err)
		return
	}
	defer texture.Destroy()

	_, _, textWidth, textHeight, err := texture.Query()
	if err != nil {
		fmt.Println("Error querying button texture:", err)
		return
	}

	textX := x + (width-textWidth)/2
	textY := buttonY + (height-textHeight)/2
	c.renderer.Copy(texture, nil, &sdl.Rect{X: textX, Y: textY, W: textWidth, H: textHeight})
}

func NewBlockingConfirmation(message string, options ...ConfirmOption) (bool, error) {
	window := internal.GetWindow()
	renderer := window.Renderer

	conf := NewConfirmationScreen(message)

	for _, option := range options {
		option(conf)
	}

	running := true
	var result bool
	var err error

	conf.Show()

	for running {
		start := time.Now()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
				err = sdl.GetError()
				return false, err

			default:
				conf.HandleEvent(event)
			}

			if conf.resultSet {
				result = conf.result
				running = false
				break
			}
		}

		deltaTime := float64(time.Since(start).Milliseconds()) / 1000.0
		conf.Update(deltaTime)

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()
		conf.Render()
		renderer.Present()

		sdl.Delay(16) // ~60fps
	}

	return result, err
}

type ConfirmOption func(*ConfirmationScreen)

func WithYesNoText(yesText, noText string) ConfirmOption {
	return func(c *ConfirmationScreen) {
		c.SetOptions(yesText, noText)
	}
}

func WithDefaultSelection(yesSelected bool) ConfirmOption {
	return func(c *ConfirmationScreen) {
		c.yesSelected = yesSelected
	}
}
