package gabagool

import (
	"math"

	"github.com/UncleJunVIP/gabagool/v2/pkg/gabagool/internal"
	"github.com/veandco/go-sdl2/sdl"
)

// ColorPicker represents a grid-based, color picker UI component
type ColorPicker struct {
	X, Y            int32
	Size            int32
	CellSize        int32
	CellPadding     int32
	GridRows        int32
	GridCols        int32
	SelectedIndex   int
	Visible         bool
	Colors          []sdl.Color
	OnColorSelected func(sdl.Color)
}

func NewHexColorPicker(window *internal.Window) *ColorPicker {
	// Center on screen
	x := window.GetWidth() / 2
	y := window.GetHeight() / 2
	size := int32(math.Min(float64(window.GetWidth()), float64(window.GetHeight())) * 0.8) // 80% of screen

	// Define grid dimensions
	gridRows := int32(5)
	gridCols := int32(5)
	cellSize := size / (int32(math.Max(float64(gridRows), float64(gridCols))) + 1)
	cellPadding := int32(4)

	// Initialize with 25 bold, highly distinguishable colors
	colors := []sdl.Color{
		// Row 1: Primary colors and variants
		{R: 255, G: 0, B: 0, A: 255},   // Red
		{R: 0, G: 255, B: 0, A: 255},   // Green
		{R: 0, G: 0, B: 255, A: 255},   // Blue
		{R: 255, G: 255, B: 0, A: 255}, // Yellow
		{R: 255, G: 0, B: 255, A: 255}, // Magenta

		// Row 2: Secondary colors and variants
		{R: 0, G: 255, B: 255, A: 255}, // Cyan
		{R: 255, G: 128, B: 0, A: 255}, // Orange
		{R: 128, G: 0, B: 255, A: 255}, // Purple
		{R: 0, G: 128, B: 0, A: 255},   // Dark Green
		{R: 128, G: 0, B: 0, A: 255},   // Maroon

		// Row 3: Tertiary colors
		{R: 0, G: 0, B: 128, A: 255},     // Navy
		{R: 0, G: 128, B: 128, A: 255},   // Teal
		{R: 128, G: 128, B: 0, A: 255},   // Olive
		{R: 128, G: 0, B: 128, A: 255},   // Purple
		{R: 255, G: 128, B: 128, A: 255}, // Pink

		// Row 4: Bright variants
		{R: 255, G: 192, B: 0, A: 255},   // Gold
		{R: 128, G: 255, B: 0, A: 255},   // Lime
		{R: 0, G: 128, B: 255, A: 255},   // Sky Blue
		{R: 255, G: 0, B: 128, A: 255},   // Rose
		{R: 128, G: 255, B: 255, A: 255}, // Light Cyan

		// Row 5: Grayscale + special colors
		{R: 255, G: 255, B: 255, A: 255}, // White
		{R: 192, G: 192, B: 192, A: 255}, // Silver
		{R: 128, G: 128, B: 128, A: 255}, // Gray
		{R: 64, G: 64, B: 64, A: 255},    // Dark Gray
		{R: 0, G: 0, B: 0, A: 255},       // Black
	}

	return &ColorPicker{
		X:               x,
		Y:               y,
		Size:            size,
		CellSize:        cellSize,
		CellPadding:     cellPadding,
		GridRows:        gridRows,
		GridCols:        gridCols,
		SelectedIndex:   0,
		Visible:         true,
		Colors:          colors,
		OnColorSelected: nil,
	}
}

func (h *ColorPicker) draw(renderer *sdl.Renderer) {
	if !h.Visible {
		return
	}

	startX := h.X - (h.GridCols*(h.CellSize+h.CellPadding))/2
	startY := h.Y - (h.GridRows*(h.CellSize+h.CellPadding))/2

	bgRect := sdl.Rect{
		X: startX - h.CellPadding,
		Y: startY - h.CellPadding,
		W: h.GridCols*(h.CellSize+h.CellPadding) + h.CellPadding,
		H: h.GridRows*(h.CellSize+h.CellPadding) + h.CellPadding,
	}
	renderer.SetDrawColor(
		255,
		255,
		255,
		255,
	)
	renderer.FillRect(&bgRect)

	borderRect := sdl.Rect{
		X: bgRect.X - 2,
		Y: bgRect.Y - 2,
		W: bgRect.W + 4,
		H: bgRect.H + 4,
	}
	renderer.SetDrawColor(
		internal.GetTheme().PrimaryAccentColor.R,
		internal.GetTheme().PrimaryAccentColor.G,
		internal.GetTheme().PrimaryAccentColor.B,
		internal.GetTheme().PrimaryAccentColor.A,
	)
	renderer.DrawRect(&borderRect)

	for i := 0; i < len(h.Colors); i++ {
		if i >= int(h.GridRows*h.GridCols) {
			break
		}

		row := int32(i) / h.GridCols
		col := int32(i) % h.GridCols

		cellX := startX + col*(h.CellSize+h.CellPadding)
		cellY := startY + row*(h.CellSize+h.CellPadding)

		cellRect := sdl.Rect{
			X: cellX,
			Y: cellY,
			W: h.CellSize,
			H: h.CellSize,
		}

		color := h.Colors[i]
		renderer.SetDrawColor(color.R, color.G, color.B, color.A)
		renderer.FillRect(&cellRect)

		if i == h.SelectedIndex {
			highlightRect := sdl.Rect{
				X: cellX - 3,
				Y: cellY - 3,
				W: h.CellSize + 6,
				H: h.CellSize + 6,
			}

			renderer.SetDrawColor(0, 0, 0, 255)
			renderer.FillRect(&highlightRect)

			cellRect := sdl.Rect{
				X: cellX,
				Y: cellY,
				W: h.CellSize,
				H: h.CellSize,
			}

			color := h.Colors[i]
			renderer.SetDrawColor(color.R, color.G, color.B, color.A)
			renderer.FillRect(&cellRect)
		} else {
			cellRect := sdl.Rect{
				X: cellX,
				Y: cellY,
				W: h.CellSize,
				H: h.CellSize,
			}

			color := h.Colors[i]
			renderer.SetDrawColor(color.R, color.G, color.B, color.A)
			renderer.FillRect(&cellRect)
		}
	}
}

func (h *ColorPicker) handleKeyPress(key sdl.Keycode) bool {
	switch key {
	case sdl.K_RIGHT, sdl.K_d:
		h.SelectedIndex = (h.SelectedIndex + 1) % len(h.Colors)
		return true

	case sdl.K_LEFT, sdl.K_a:
		h.SelectedIndex = (h.SelectedIndex - 1 + len(h.Colors)) % len(h.Colors)
		return true

	case sdl.K_UP, sdl.K_w:
		if h.SelectedIndex >= int(h.GridCols) {
			h.SelectedIndex -= int(h.GridCols)
		} else {
			lastRowStart := ((len(h.Colors) - 1) / int(h.GridCols)) * int(h.GridCols)
			h.SelectedIndex = int(math.Min(float64(lastRowStart+h.SelectedIndex), float64(len(h.Colors)-1)))
		}
		return true

	case sdl.K_DOWN, sdl.K_s:
		if h.SelectedIndex+int(h.GridCols) < len(h.Colors) {
			h.SelectedIndex += int(h.GridCols)
		} else {
			h.SelectedIndex = h.SelectedIndex % int(h.GridCols)
		}
		return true

	case sdl.K_RETURN, sdl.K_SPACE:
		if h.OnColorSelected != nil {
			h.OnColorSelected(h.Colors[h.SelectedIndex])
		}
		return true
	}

	return false
}

func (h *ColorPicker) setVisible(visible bool) {
	h.Visible = visible
}

func (h *ColorPicker) getSelectedColor() sdl.Color {
	if h.SelectedIndex >= 0 && h.SelectedIndex < len(h.Colors) {
		return h.Colors[h.SelectedIndex]
	}
	return sdl.Color{R: 255, G: 255, B: 255, A: 255}
}

func (h *ColorPicker) setOnColorSelected(callback func(sdl.Color)) {
	h.OnColorSelected = callback
}
