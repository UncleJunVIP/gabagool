package gabagool

import (
	"github.com/veandco/go-sdl2/sdl"
	"math"
)

// ColorPicker represents a color picker UI element (now grid-based)
type ColorPicker struct {
	// Position and size
	X, Y            int32           // Center position
	Size            int32           // Size of the picker
	CellSize        int32           // Size of individual color cells
	CellPadding     int32           // Padding between cells
	GridRows        int32           // Number of rows in the grid
	GridCols        int32           // Number of columns in the grid
	SelectedIndex   int             // Currently selected color cell
	Visible         bool            // Is the picker visible?
	Colors          []sdl.Color     // Available colors in the picker
	OnColorSelected func(sdl.Color) // Callback when a color is selected
}

// NewHexColorPicker creates a new grid-based color picker centered on screen
func NewHexColorPicker(window *Window) *ColorPicker {
	// Center on screen
	x := window.Width / 2
	y := window.Height / 2
	size := int32(math.Min(float64(window.Width), float64(window.Height)) * 0.8) // 80% of screen

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

// Draw renders the grid-based color picker
func (h *ColorPicker) Draw(renderer *sdl.Renderer) {
	if !h.Visible {
		return
	}

	// Calculate the top-left corner of the grid
	startX := h.X - (h.GridCols*(h.CellSize+h.CellPadding))/2
	startY := h.Y - (h.GridRows*(h.CellSize+h.CellPadding))/2

	// Draw background rectangle
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

	// Draw border
	borderRect := sdl.Rect{
		X: bgRect.X - 2,
		Y: bgRect.Y - 2,
		W: bgRect.W + 4,
		H: bgRect.H + 4,
	}
	renderer.SetDrawColor(
		GetTheme().PrimaryAccentColor.R,
		GetTheme().PrimaryAccentColor.G,
		GetTheme().PrimaryAccentColor.B,
		GetTheme().PrimaryAccentColor.A,
	)
	renderer.DrawRect(&borderRect)

	// Draw color cells
	for i := 0; i < len(h.Colors); i++ {
		if i >= int(h.GridRows*h.GridCols) {
			break // Don't draw more cells than can fit in the grid
		}

		row := int32(i) / h.GridCols
		col := int32(i) % h.GridCols

		cellX := startX + col*(h.CellSize+h.CellPadding)
		cellY := startY + row*(h.CellSize+h.CellPadding)

		// Draw the color cell
		cellRect := sdl.Rect{
			X: cellX,
			Y: cellY,
			W: h.CellSize,
			H: h.CellSize,
		}

		// Set color
		color := h.Colors[i]
		renderer.SetDrawColor(color.R, color.G, color.B, color.A)
		renderer.FillRect(&cellRect)

		// Draw selection indicator
		if i == h.SelectedIndex {
			// Define the highlight rect to completely enclose the color cell
			highlightRect := sdl.Rect{
				X: cellX - 3,
				Y: cellY - 3,
				W: h.CellSize + 6,
				H: h.CellSize + 6,
			}

			renderer.SetDrawColor(0, 0, 0, 255) // Bright yellow
			renderer.FillRect(&highlightRect)

			// Now redraw the color cell on top, but slightly smaller to create a visible border
			cellRect := sdl.Rect{
				X: cellX,
				Y: cellY,
				W: h.CellSize,
				H: h.CellSize,
			}

			// Set color
			color := h.Colors[i]
			renderer.SetDrawColor(color.R, color.G, color.B, color.A)
			renderer.FillRect(&cellRect)
		} else {
			// For non-selected cells, just draw the color normally
			cellRect := sdl.Rect{
				X: cellX,
				Y: cellY,
				W: h.CellSize,
				H: h.CellSize,
			}

			// Set color
			color := h.Colors[i]
			renderer.SetDrawColor(color.R, color.G, color.B, color.A)
			renderer.FillRect(&cellRect)
		}
	}
}

// HandleEvent processes input events for the color picker
func (h *ColorPicker) HandleEvent(event sdl.Event) bool {
	if !h.Visible {
		return false
	}

	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		if e.Type == sdl.KEYDOWN {
			return h.handleKeyPress(e.Keysym.Sym)
		}
	}

	return false
}

// handleKeyPress handles keyboard navigation of the color picker
func (h *ColorPicker) handleKeyPress(key sdl.Keycode) bool {
	// Arrow key navigation
	switch key {
	case sdl.K_RIGHT, sdl.K_d:
		h.SelectedIndex = (h.SelectedIndex + 1) % len(h.Colors)
		return true

	case sdl.K_LEFT, sdl.K_a:
		h.SelectedIndex = (h.SelectedIndex - 1 + len(h.Colors)) % len(h.Colors)
		return true

	case sdl.K_UP, sdl.K_w:
		// Move up one row
		if h.SelectedIndex >= int(h.GridCols) {
			h.SelectedIndex -= int(h.GridCols)
		} else {
			// Wrap to the last row
			lastRowStart := ((len(h.Colors) - 1) / int(h.GridCols)) * int(h.GridCols)
			h.SelectedIndex = int(math.Min(float64(lastRowStart+h.SelectedIndex), float64(len(h.Colors)-1)))
		}
		return true

	case sdl.K_DOWN, sdl.K_s:
		// Move down one row
		if h.SelectedIndex+int(h.GridCols) < len(h.Colors) {
			h.SelectedIndex += int(h.GridCols)
		} else {
			// Wrap to the first row
			h.SelectedIndex = h.SelectedIndex % int(h.GridCols)
		}
		return true

	case sdl.K_RETURN, sdl.K_SPACE:
		// Select the current color
		if h.OnColorSelected != nil {
			h.OnColorSelected(h.Colors[h.SelectedIndex])
		}
		return true
	}

	return false
}

// SetVisible shows or hides the color picker
func (h *ColorPicker) SetVisible(visible bool) {
	h.Visible = visible
}

// GetSelectedColor returns the currently selected color
func (h *ColorPicker) GetSelectedColor() sdl.Color {
	if h.SelectedIndex >= 0 && h.SelectedIndex < len(h.Colors) {
		return h.Colors[h.SelectedIndex]
	}
	return sdl.Color{R: 255, G: 255, B: 255, A: 255}
}

// SetColors allows custom colors to be set for the picker
func (h *ColorPicker) SetColors(colors []sdl.Color) {
	h.Colors = colors
	if h.SelectedIndex >= len(h.Colors) {
		h.SelectedIndex = 0
	}
}

// SetOnColorSelected sets the callback function for when a color is selected
func (h *ColorPicker) SetOnColorSelected(callback func(sdl.Color)) {
	h.OnColorSelected = callback
}

// InitColors initializes the color picker with 25 bold, distinct colors for retro gaming
func (h *ColorPicker) InitColors() {
	// Set grid dimensions to accommodate 25 colors (5x5 grid)
	h.GridRows = 5
	h.GridCols = 5

	// Initialize with 25 bold, highly distinguishable colors
	h.Colors = []sdl.Color{
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
}
