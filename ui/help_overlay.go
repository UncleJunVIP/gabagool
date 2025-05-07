package ui

import (
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// HelpOverlay represents a help screen that can be toggled in UI components
type HelpOverlay struct {
	Lines           []string
	ShowingHelp     bool
	ScrollOffset    int32
	MaxScrollOffset int32
	Padding         int32
	LineHeight      int32
	Width           int32
	Height          int32
	BackgroundColor sdl.Color
	TextColor       sdl.Color
	ScrollbarColor  sdl.Color
	ScrollbarWidth  int32
	ExitTextPadding int32 // Additional padding for exit text
}

// NewHelpOverlay creates a new help overlay with default settings
func NewHelpOverlay(lines []string) *HelpOverlay {
	window := internal.GetWindow()
	width, height := window.Window.GetSize()

	return &HelpOverlay{
		Lines:           lines,
		ShowingHelp:     false,
		ScrollOffset:    0,
		MaxScrollOffset: 0,
		Padding:         20,
		LineHeight:      50,
		Width:           width,
		Height:          height,
		BackgroundColor: sdl.Color{R: 0, G: 0, B: 0, A: 220},       // Semi-transparent black
		TextColor:       sdl.Color{R: 255, G: 255, B: 255, A: 255}, // White
		ScrollbarColor:  sdl.Color{R: 150, G: 150, B: 150, A: 180}, // Semi-transparent gray
		ScrollbarWidth:  8,
		ExitTextPadding: 40, // Increased padding for exit text
	}
}

// Toggle shows or hides the help overlay
func (h *HelpOverlay) Toggle() {
	h.ShowingHelp = !h.ShowingHelp
	if h.ShowingHelp {
		h.ScrollOffset = 0 // Reset scroll position when showing help
	}
}

// Scroll moves the help content up or down
func (h *HelpOverlay) Scroll(direction int) {
	if !h.ShowingHelp {
		return
	}

	// Calculate how much we can scroll based on content height
	newOffset := h.ScrollOffset + int32(direction)*h.LineHeight

	// Prevent scrolling past the beginning
	if newOffset < 0 {
		newOffset = 0
	}

	// Prevent scrolling past the end
	if newOffset > h.MaxScrollOffset {
		newOffset = h.MaxScrollOffset
	}

	h.ScrollOffset = newOffset
}

// CalculateMaxScroll calculates how far down the help text can be scrolled
func (h *HelpOverlay) CalculateMaxScroll(renderer *sdl.Renderer, font *ttf.Font) {
	totalHeight := int32(len(h.Lines)) * h.LineHeight
	visibleHeight := h.Height - (h.Padding * 2) - h.LineHeight - h.ExitTextPadding // Reserve space for exit instructions with padding

	if totalHeight > visibleHeight {
		h.MaxScrollOffset = totalHeight - visibleHeight
	} else {
		h.MaxScrollOffset = 0
	}
}

// Render draws the help overlay on the screen
func (h *HelpOverlay) Render(renderer *sdl.Renderer, font *ttf.Font) {
	if !h.ShowingHelp {
		return
	}

	// Ensure max scroll calculation is up to date
	h.CalculateMaxScroll(renderer, font)

	// Draw semi-transparent background
	bgRect := &sdl.Rect{X: 0, Y: 0, W: h.Width, H: h.Height}
	renderer.SetDrawColor(h.BackgroundColor.R, h.BackgroundColor.G, h.BackgroundColor.B, h.BackgroundColor.A)
	renderer.FillRect(bgRect)

	// Draw title
	titleText, err := font.RenderUTF8Blended("Help", h.TextColor)
	if err == nil {
		titleTexture, err := renderer.CreateTextureFromSurface(titleText)
		if err == nil {
			titleRect := &sdl.Rect{
				X: h.Padding,
				Y: h.Padding,
				W: titleText.W,
				H: titleText.H,
			}
			renderer.Copy(titleTexture, nil, titleRect)
			titleTexture.Destroy()
		}
		titleText.Free()
	}

	// Calculate content area
	contentY := h.Padding + h.LineHeight*2
	contentHeight := h.Height - contentY - h.Padding - h.LineHeight - h.ExitTextPadding // Reserve more space at bottom for exit text

	// Calculate content width (accounting for scrollbar)
	contentWidth := h.Width - h.Padding*2
	if h.MaxScrollOffset > 0 {
		contentWidth -= (h.ScrollbarWidth + 10) // Add some spacing between content and scrollbar
	}

	// Draw help lines
	y := contentY - h.ScrollOffset

	for _, line := range h.Lines {
		// Skip lines that would be rendered outside the visible area
		if y+h.LineHeight < contentY || y > contentY+contentHeight {
			y += h.LineHeight
			continue
		}

		textSurface, err := font.RenderUTF8Blended(line, h.TextColor)
		if err == nil {
			textTexture, err := renderer.CreateTextureFromSurface(textSurface)
			if err == nil {
				textRect := &sdl.Rect{
					X: h.Padding,
					Y: y,
					W: textSurface.W,
					H: textSurface.H,
				}
				renderer.Copy(textTexture, nil, textRect)
				textTexture.Destroy()
			}
			textSurface.Free()
		}

		y += h.LineHeight
	}

	// Draw scrollbar if needed
	if h.MaxScrollOffset > 0 {
		// Calculate scrollbar positioning
		scrollbarX := h.Width - h.Padding - h.ScrollbarWidth
		scrollbarY := contentY
		scrollbarHeight := contentHeight

		// Draw scrollbar background/track
		scrollbarBgRect := &sdl.Rect{
			X: scrollbarX,
			Y: scrollbarY,
			W: h.ScrollbarWidth,
			H: scrollbarHeight,
		}
		scrollbarBgColor := sdl.Color{R: 50, G: 50, B: 50, A: 100}
		renderer.SetDrawColor(scrollbarBgColor.R, scrollbarBgColor.G, scrollbarBgColor.B, scrollbarBgColor.A)
		renderer.FillRect(scrollbarBgRect)

		// Calculate scrollbar handle size and position
		totalContentHeight := int32(len(h.Lines)) * h.LineHeight
		handleRatio := float64(contentHeight) / float64(totalContentHeight)
		if handleRatio > 1.0 {
			handleRatio = 1.0
		}

		handleHeight := int32(float64(scrollbarHeight) * handleRatio)
		if handleHeight < 30 {
			handleHeight = 30 // Minimum handle size
		}

		scrollRatio := 0.0
		if h.MaxScrollOffset > 0 {
			scrollRatio = float64(h.ScrollOffset) / float64(h.MaxScrollOffset)
		}

		handleY := scrollbarY + int32(float64(scrollbarHeight-handleHeight)*scrollRatio)

		// Draw scrollbar handle
		scrollbarHandleRect := &sdl.Rect{
			X: scrollbarX,
			Y: handleY,
			W: h.ScrollbarWidth,
			H: handleHeight,
		}
		renderer.SetDrawColor(h.ScrollbarColor.R, h.ScrollbarColor.G, h.ScrollbarColor.B, h.ScrollbarColor.A)
		renderer.FillRect(scrollbarHandleRect)
	}

	// Draw exit instructions with increased padding
	exitText := "Press any button to close help"
	exitSurface, err := font.RenderUTF8Blended(exitText, h.TextColor)
	if err == nil {
		exitTexture, err := renderer.CreateTextureFromSurface(exitSurface)
		if err == nil {
			exitRect := &sdl.Rect{
				X: (h.Width - exitSurface.W) / 2,
				Y: h.Height - h.Padding - exitSurface.H - h.ExitTextPadding/2, // Add half of the exit text padding
				W: exitSurface.W,
				H: exitSurface.H,
			}
			renderer.Copy(exitTexture, nil, exitRect)
			exitTexture.Destroy()
		}
		exitSurface.Free()
	}
}
