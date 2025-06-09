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
	Visible         bool            // Is the picker visible
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
	gridRows := int32(7)
	gridCols := int32(7)
	cellSize := size / (int32(math.Max(float64(gridRows), float64(gridCols))) + 1)
	cellPadding := int32(4)

	// Create an extended color palette with at least 48 colors
	colors := []sdl.Color{
		// Reds
		{R: 255, G: 0, B: 0, A: 255},     // Red
		{R: 220, G: 20, B: 60, A: 255},   // Crimson
		{R: 178, G: 34, B: 34, A: 255},   // Firebrick
		{R: 139, G: 0, B: 0, A: 255},     // Dark red
		{R: 255, G: 99, B: 71, A: 255},   // Tomato
		{R: 205, G: 92, B: 92, A: 255},   // Indian red
		{R: 240, G: 128, B: 128, A: 255}, // Light coral

		// Oranges
		{R: 255, G: 127, B: 0, A: 255}, // Orange
		{R: 255, G: 140, B: 0, A: 255}, // Dark orange
		{R: 255, G: 165, B: 0, A: 255}, // Orange
		{R: 255, G: 69, B: 0, A: 255},  // Orange red

		// Yellows
		{R: 255, G: 255, B: 0, A: 255},   // Yellow
		{R: 255, G: 215, B: 0, A: 255},   // Gold
		{R: 218, G: 165, B: 32, A: 255},  // Goldenrod
		{R: 240, G: 230, B: 140, A: 255}, // Khaki

		// Greens
		{R: 0, G: 128, B: 0, A: 255},     // Green
		{R: 34, G: 139, B: 34, A: 255},   // Forest green
		{R: 0, G: 255, B: 0, A: 255},     // Lime
		{R: 50, G: 205, B: 50, A: 255},   // Lime green
		{R: 144, G: 238, B: 144, A: 255}, // Light green
		{R: 152, G: 251, B: 152, A: 255}, // Pale green
		{R: 143, G: 188, B: 143, A: 255}, // Dark sea green

		// Cyans
		{R: 0, G: 255, B: 255, A: 255},  // Cyan
		{R: 0, G: 139, B: 139, A: 255},  // Dark cyan
		{R: 32, G: 178, B: 170, A: 255}, // Light sea green
		{R: 64, G: 224, B: 208, A: 255}, // Turquoise

		// Blues
		{R: 0, G: 0, B: 255, A: 255},     // Blue
		{R: 0, G: 0, B: 139, A: 255},     // Dark blue
		{R: 0, G: 0, B: 205, A: 255},     // Medium blue
		{R: 65, G: 105, B: 225, A: 255},  // Royal blue
		{R: 100, G: 149, B: 237, A: 255}, // Cornflower blue
		{R: 135, G: 206, B: 235, A: 255}, // Sky blue
		{R: 135, G: 206, B: 250, A: 255}, // Light sky blue

		// Purples
		{R: 128, G: 0, B: 128, A: 255},  // Purple
		{R: 148, G: 0, B: 211, A: 255},  // Dark violet
		{R: 153, G: 50, B: 204, A: 255}, // Dark orchid
		{R: 186, G: 85, B: 211, A: 255}, // Medium orchid
		{R: 138, G: 43, B: 226, A: 255}, // Blue violet
		{R: 75, G: 0, B: 130, A: 255},   // Indigo

		// Pinks
		{R: 255, G: 192, B: 203, A: 255}, // Pink
		{R: 255, G: 105, B: 180, A: 255}, // Hot pink
		{R: 219, G: 112, B: 147, A: 255}, // Pale violet red
		{R: 199, G: 21, B: 133, A: 255},  // Medium violet red

		// Browns
		{R: 165, G: 42, B: 42, A: 255},  // Brown
		{R: 139, G: 69, B: 19, A: 255},  // Saddle brown
		{R: 160, G: 82, B: 45, A: 255},  // Sienna
		{R: 210, G: 105, B: 30, A: 255}, // Chocolate

		// Neutrals
		{R: 255, G: 255, B: 255, A: 255}, // White
		{R: 192, G: 192, B: 192, A: 255}, // Silver
		{R: 128, G: 128, B: 128, A: 255}, // Gray
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

func (h *ColorPicker) InitColors() {
	// Initialize with the specific hex colors provided
	h.Colors = []sdl.Color{
		// Blues
		{R: 0x00, G: 0x00, B: 0x11, A: 255},
		{R: 0x00, G: 0x00, B: 0x22, A: 255},
		{R: 0x00, G: 0x00, B: 0x33, A: 255},
		{R: 0x00, G: 0x00, B: 0x44, A: 255},
		{R: 0x00, G: 0x00, B: 0x55, A: 255},
		{R: 0x00, G: 0x00, B: 0x66, A: 255},
		{R: 0x00, G: 0x00, B: 0x77, A: 255},
		{R: 0x00, G: 0x00, B: 0x88, A: 255},
		{R: 0x00, G: 0x00, B: 0x99, A: 255},
		{R: 0x00, G: 0x00, B: 0xAA, A: 255},
		{R: 0x00, G: 0x00, B: 0xBB, A: 255},
		{R: 0x00, G: 0x00, B: 0xCC, A: 255},
		{R: 0x33, G: 0x66, B: 0xFF, A: 255},
		{R: 0x4D, G: 0x7A, B: 0xFF, A: 255},
		{R: 0x66, G: 0x99, B: 0xFF, A: 255},
		{R: 0x80, G: 0xB3, B: 0xFF, A: 255},
		{R: 0x99, G: 0xCC, B: 0xFF, A: 255},
		{R: 0xB3, G: 0xD9, B: 0xFF, A: 255},
		{R: 0x00, G: 0x00, B: 0xFF, A: 255},

		// Cyan
		{R: 0x00, G: 0x11, B: 0x11, A: 255},
		{R: 0x00, G: 0x22, B: 0x22, A: 255},
		{R: 0x00, G: 0x33, B: 0x33, A: 255},
		{R: 0x00, G: 0x44, B: 0x44, A: 255},
		{R: 0x00, G: 0x55, B: 0x55, A: 255},
		{R: 0x00, G: 0x66, B: 0x66, A: 255},
		{R: 0x00, G: 0x77, B: 0x77, A: 255},
		{R: 0x00, G: 0x88, B: 0x88, A: 255},
		{R: 0x00, G: 0x99, B: 0x99, A: 255},
		{R: 0x00, G: 0xAA, B: 0xAA, A: 255},
		{R: 0x00, G: 0xBB, B: 0xBB, A: 255},
		{R: 0x00, G: 0xCC, B: 0xCC, A: 255},
		{R: 0x33, G: 0xFF, B: 0xFF, A: 255},
		{R: 0x4D, G: 0xFF, B: 0xFF, A: 255},
		{R: 0x66, G: 0xFF, B: 0xFF, A: 255},
		{R: 0x80, G: 0xFF, B: 0xFF, A: 255},
		{R: 0x99, G: 0xFF, B: 0xFF, A: 255},
		{R: 0xB3, G: 0xFF, B: 0xFF, A: 255},
		{R: 0x00, G: 0xFF, B: 0xFF, A: 255},

		// Green
		{R: 0x00, G: 0x11, B: 0x00, A: 255},
		{R: 0x00, G: 0x22, B: 0x00, A: 255},
		{R: 0x00, G: 0x33, B: 0x00, A: 255},
		{R: 0x00, G: 0x44, B: 0x00, A: 255},
		{R: 0x00, G: 0x55, B: 0x00, A: 255},
		{R: 0x00, G: 0x66, B: 0x00, A: 255},
		{R: 0x00, G: 0x77, B: 0x00, A: 255},
		{R: 0x00, G: 0x88, B: 0x00, A: 255},
		{R: 0x00, G: 0x99, B: 0x00, A: 255},
		{R: 0x00, G: 0xAA, B: 0x00, A: 255},
		{R: 0x00, G: 0xBB, B: 0x00, A: 255},
		{R: 0x00, G: 0xCC, B: 0x00, A: 255},
		{R: 0x33, G: 0xFF, B: 0x33, A: 255},
		{R: 0x4D, G: 0xFF, B: 0x4D, A: 255},
		{R: 0x66, G: 0xFF, B: 0x66, A: 255},
		{R: 0x80, G: 0xFF, B: 0x80, A: 255},
		{R: 0x99, G: 0xFF, B: 0x99, A: 255},
		{R: 0xB3, G: 0xFF, B: 0xB3, A: 255},
		{R: 0x00, G: 0xFF, B: 0x00, A: 255},

		// Magenta
		{R: 0x11, G: 0x00, B: 0x11, A: 255},
		{R: 0x22, G: 0x00, B: 0x22, A: 255},
		{R: 0x33, G: 0x00, B: 0x33, A: 255},
		{R: 0x44, G: 0x00, B: 0x44, A: 255},
		{R: 0x55, G: 0x00, B: 0x55, A: 255},
		{R: 0x66, G: 0x00, B: 0x66, A: 255},
		{R: 0x77, G: 0x00, B: 0x77, A: 255},
		{R: 0x88, G: 0x00, B: 0x88, A: 255},
		{R: 0x99, G: 0x00, B: 0x99, A: 255},
		{R: 0xAA, G: 0x00, B: 0xAA, A: 255},
		{R: 0xBB, G: 0x00, B: 0xBB, A: 255},
		{R: 0xCC, G: 0x00, B: 0xCC, A: 255},
		{R: 0xFF, G: 0x33, B: 0xFF, A: 255},
		{R: 0xFF, G: 0x4D, B: 0xFF, A: 255},
		{R: 0xFF, G: 0x66, B: 0xFF, A: 255},
		{R: 0xFF, G: 0x80, B: 0xFF, A: 255},
		{R: 0xFF, G: 0x99, B: 0xFF, A: 255},
		{R: 0xFF, G: 0xB3, B: 0xFF, A: 255},
		{R: 0xFF, G: 0x00, B: 0xFF, A: 255},

		// Purple
		{R: 0x22, G: 0x00, B: 0x44, A: 255},
		{R: 0x33, G: 0x00, B: 0x66, A: 255},
		{R: 0x44, G: 0x00, B: 0x88, A: 255},
		{R: 0x55, G: 0x00, B: 0xAA, A: 255},
		{R: 0x66, G: 0x00, B: 0xCC, A: 255},
		{R: 0x77, G: 0x00, B: 0xDD, A: 255},
		{R: 0x88, G: 0x00, B: 0xEE, A: 255},
		{R: 0x99, G: 0x00, B: 0xFF, A: 255},
		{R: 0xAA, G: 0x00, B: 0xFF, A: 255},
		{R: 0xBB, G: 0x00, B: 0xFF, A: 255},
		{R: 0xCC, G: 0x00, B: 0xFF, A: 255},
		{R: 0x88, G: 0x33, B: 0xFF, A: 255},
		{R: 0x99, G: 0x4D, B: 0xFF, A: 255},
		{R: 0xAA, G: 0x66, B: 0xFF, A: 255},
		{R: 0xBB, G: 0x80, B: 0xFF, A: 255},
		{R: 0xCC, G: 0x99, B: 0xFF, A: 255},
		{R: 0xDD, G: 0xB3, B: 0xFF, A: 255},

		// Red
		{R: 0x22, G: 0x00, B: 0x00, A: 255},
		{R: 0x44, G: 0x00, B: 0x00, A: 255},
		{R: 0x66, G: 0x00, B: 0x00, A: 255},
		{R: 0x88, G: 0x00, B: 0x00, A: 255},
		{R: 0xAA, G: 0x00, B: 0x00, A: 255},
		{R: 0xCC, G: 0x00, B: 0x00, A: 255},
		{R: 0xFF, G: 0x33, B: 0x33, A: 255},
		{R: 0xFF, G: 0x4D, B: 0x4D, A: 255},
		{R: 0xFF, G: 0x66, B: 0x66, A: 255},
		{R: 0xFF, G: 0x80, B: 0x80, A: 255},
		{R: 0xFF, G: 0x99, B: 0x99, A: 255},
		{R: 0xFF, G: 0xB3, B: 0xB3, A: 255},
		{R: 0xFF, G: 0x00, B: 0x00, A: 255},

		// Yellow
		{R: 0x22, G: 0x22, B: 0x00, A: 255},
		{R: 0x44, G: 0x44, B: 0x00, A: 255},
		{R: 0x66, G: 0x66, B: 0x00, A: 255},
		{R: 0x88, G: 0x88, B: 0x00, A: 255},
		{R: 0xAA, G: 0xAA, B: 0x00, A: 255},
		{R: 0xCC, G: 0xCC, B: 0x00, A: 255},
		{R: 0xFF, G: 0xFF, B: 0x33, A: 255},
		{R: 0xFF, G: 0xFF, B: 0x4D, A: 255},
		{R: 0xFF, G: 0xFF, B: 0x66, A: 255},
		{R: 0xFF, G: 0xFF, B: 0x80, A: 255},
		{R: 0xFF, G: 0xFF, B: 0x99, A: 255},
		{R: 0xFF, G: 0xFF, B: 0xB3, A: 255},
		{R: 0xFF, G: 0xFF, B: 0x00, A: 255},

		// Orange
		{R: 0x33, G: 0x11, B: 0x00, A: 255},
		{R: 0x66, G: 0x22, B: 0x00, A: 255},
		{R: 0x99, G: 0x33, B: 0x00, A: 255},
		{R: 0xCC, G: 0x44, B: 0x00, A: 255},
		{R: 0xFF, G: 0x55, B: 0x00, A: 255},
		{R: 0xFF, G: 0x66, B: 0x00, A: 255},
		{R: 0xFF, G: 0x77, B: 0x11, A: 255},
		{R: 0xFF, G: 0x88, B: 0x22, A: 255},
		{R: 0xFF, G: 0x99, B: 0x33, A: 255},
		{R: 0xFF, G: 0xAA, B: 0x44, A: 255},
		{R: 0xFF, G: 0xBB, B: 0x55, A: 255},
		{R: 0xFF, G: 0xCC, B: 0x66, A: 255},
		{R: 0xFF, G: 0xDD, B: 0x77, A: 255},
		{R: 0xFF, G: 0xEE, B: 0x88, A: 255},

		// White to Black Gradient
		{R: 0x00, G: 0x00, B: 0x00, A: 255},
		{R: 0x11, G: 0x11, B: 0x11, A: 255},
		{R: 0x22, G: 0x22, B: 0x22, A: 255},
		{R: 0x33, G: 0x33, B: 0x33, A: 255},
		{R: 0x44, G: 0x44, B: 0x44, A: 255},
		{R: 0x55, G: 0x55, B: 0x55, A: 255},
		{R: 0x66, G: 0x66, B: 0x66, A: 255},
		{R: 0x77, G: 0x77, B: 0x77, A: 255},
		{R: 0x88, G: 0x88, B: 0x88, A: 255},
		{R: 0x99, G: 0x99, B: 0x99, A: 255},
		{R: 0xAA, G: 0xAA, B: 0xAA, A: 255},
		{R: 0xBB, G: 0xBB, B: 0xBB, A: 255},
		{R: 0xCC, G: 0xCC, B: 0xCC, A: 255},
		{R: 0xDD, G: 0xDD, B: 0xDD, A: 255},
		{R: 0xFF, G: 0xFF, B: 0xFF, A: 255},
	}
}
