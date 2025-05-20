package gabagool

import (
	"github.com/veandco/go-sdl2/ttf"
	"strings"
	"time"

	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type MetadataItem struct {
	Label string
	Value string
}

type DetailScreenOptions struct {
	ImagePaths          []string
	Metadata            []MetadataItem
	InfoLabel           string
	Description         string
	TitleColor          sdl.Color
	MetadataColor       sdl.Color
	DescriptionColor    sdl.Color
	BackgroundColor     sdl.Color
	MaxImageHeight      int32
	MaxImageWidth       int32
	ShowScrollbar       bool
	ShowThemeBackground bool
}

type DetailScreenReturn struct {
	LastPressedKey sdl.Keycode
	LastPressedBtn uint8
	Cancelled      bool
}

func DefaultInfoScreenOptions() DetailScreenOptions {
	return DetailScreenOptions{
		ImagePaths:       []string{},
		Metadata:         []MetadataItem{},
		Description:      "",
		TitleColor:       sdl.Color{R: 255, G: 255, B: 255, A: 255},
		MetadataColor:    sdl.Color{R: 220, G: 220, B: 220, A: 255},
		DescriptionColor: sdl.Color{R: 200, G: 200, B: 200, A: 255},
		BackgroundColor:  sdl.Color{R: 0, G: 0, B: 0, A: 255},
		ShowScrollbar:    true,
	}
}

func DetailScreen(title string, options DetailScreenOptions, footerHelpItems []FooterHelpItem) (types.Option[DetailScreenReturn], error) {
	window := GetWindow()
	renderer := window.Renderer

	footerHeight := int32(30)
	safeAreaHeight := window.Height - footerHeight

	if options.MaxImageHeight == 0 {
		options.MaxImageHeight = int32(float64(safeAreaHeight) / 2)
	}
	if options.MaxImageWidth == 0 {
		options.MaxImageWidth = int32(float64(window.Width) / 2)
	}

	// Target scroll position for smooth scrolling
	targetScrollY := int32(0)
	// Current scroll position that will be animated
	scrollY := int32(0)
	maxScrollY := int32(0)
	scrollSpeed := int32(85)
	scrollbarWidth := int32(10)

	// Smooth scrolling variables
	scrollAnimationSpeed := float32(0.15) // Lower for smoother animation (0.1-0.3 is good)

	// Directional repeat tracking
	heldDirections := struct {
		up   bool
		down bool
	}{}
	lastRepeatTime := time.Now()
	repeatDelay := time.Millisecond * 100    // Initial delay before repeating
	repeatInterval := time.Millisecond * 100 // Interval between repeats

	currentImageIndex := 0
	imageTextures := []*sdl.Texture{}
	imageDimensions := []sdl.Rect{}

	if len(options.ImagePaths) > 0 {
		for _, imagePath := range options.ImagePaths {
			image, err := img.Load(imagePath)
			if err == nil {
				defer image.Free()

				texture, err := renderer.CreateTextureFromSurface(image)
				if err == nil {
					imageTextures = append(imageTextures, texture)

					imageW := image.W
					imageH := image.H

					if imageW > options.MaxImageWidth {
						ratio := float32(options.MaxImageWidth) / float32(imageW)
						imageW = options.MaxImageWidth
						imageH = int32(float32(imageH) * ratio)
					}

					if imageH > options.MaxImageHeight {
						ratio := float32(options.MaxImageHeight) / float32(imageH)
						imageH = options.MaxImageHeight
						imageW = int32(float32(imageW) * ratio)
					}

					rect := sdl.Rect{
						X: (window.Width - imageW) / 2,
						Y: 0,
						W: imageW,
						H: imageH,
					}

					imageDimensions = append(imageDimensions, rect)
				}
			}
		}
	}

	lastInputTime := time.Now()
	inputDelay := DefaultInputDelay

	result := DetailScreenReturn{
		LastPressedKey: 0,
		LastPressedBtn: 0,
		Cancelled:      true,
	}

	running := true
	firstRender := true

	// Function to handle directional repeats
	handleDirectionalRepeats := func() {
		now := time.Now()
		timeSinceLastRepeat := now.Sub(lastRepeatTime)

		// Initial delay before repeating
		if timeSinceLastRepeat < repeatDelay {
			return
		}

		// After initial delay, repeat at the specified interval
		if repeatInterval > 0 && timeSinceLastRepeat < repeatInterval {
			return
		}

		// Process held directions
		if heldDirections.up {
			targetScrollY = max(0, targetScrollY-scrollSpeed)
			lastRepeatTime = now
		} else if heldDirections.down {
			targetScrollY = min(maxScrollY, targetScrollY+scrollSpeed)
			lastRepeatTime = now
		}
	}

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				result.Cancelled = true

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					result.LastPressedKey = e.Keysym.Sym

					currentTime := time.Now()
					if currentTime.Sub(lastInputTime) < inputDelay {
						continue
					}
					lastInputTime = currentTime

					switch e.Keysym.Sym {
					case sdl.K_UP:
						heldDirections.up = true
						targetScrollY = max(0, targetScrollY-scrollSpeed)
						lastRepeatTime = currentTime
					case sdl.K_DOWN:
						heldDirections.down = true
						targetScrollY = min(maxScrollY, targetScrollY+scrollSpeed)
						lastRepeatTime = currentTime
					case sdl.K_LEFT:
						if len(imageTextures) > 1 {
							currentImageIndex = (currentImageIndex - 1 + len(imageTextures)) % len(imageTextures)
							lastRepeatTime = currentTime
						}
					case sdl.K_RIGHT:
						if len(imageTextures) > 1 {
							currentImageIndex = (currentImageIndex + 1) % len(imageTextures)
							lastRepeatTime = currentTime
						}
					case sdl.K_b, sdl.K_ESCAPE:
						result.Cancelled = true
						running = false
					case sdl.K_a, sdl.K_RETURN:
						result.Cancelled = false
						running = false
					}
				} else if e.Type == sdl.KEYUP {
					switch e.Keysym.Sym {
					case sdl.K_UP:
						heldDirections.up = false
					case sdl.K_DOWN:
						heldDirections.down = false
					}
				}

			case *sdl.ControllerButtonEvent:
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					result.LastPressedBtn = e.Button

					currentTime := time.Now()
					if currentTime.Sub(lastInputTime) < inputDelay {
						continue
					}
					lastInputTime = currentTime

					switch e.Button {
					case BrickButton_UP:
						heldDirections.up = true
						targetScrollY = max(0, targetScrollY-scrollSpeed)
						lastRepeatTime = currentTime
					case BrickButton_DOWN:
						heldDirections.down = true
						targetScrollY = min(maxScrollY, targetScrollY+scrollSpeed)
						lastRepeatTime = currentTime
					case BrickButton_LEFT:
						if len(imageTextures) > 1 {
							currentImageIndex = (currentImageIndex - 1 + len(imageTextures)) % len(imageTextures)
							lastRepeatTime = currentTime
						}
					case BrickButton_RIGHT:
						if len(imageTextures) > 1 {
							currentImageIndex = (currentImageIndex + 1) % len(imageTextures)
							lastRepeatTime = currentTime
						}
					case BrickButton_B:
						result.Cancelled = true
						running = false
					case BrickButton_A:
						result.Cancelled = false
						running = false
					}
				} else if e.Type == sdl.CONTROLLERBUTTONUP {
					switch e.Button {
					case BrickButton_UP:
						heldDirections.up = false
					case BrickButton_DOWN:
						heldDirections.down = false
					}
				}
			}
		}

		// Handle directional repeats
		handleDirectionalRepeats()

		// Smooth scrolling animation - interpolate between current position and target
		scrollY += int32(float32(targetScrollY-scrollY) * scrollAnimationSpeed)

		// Set background color - FIXED
		if options.ShowThemeBackground {
			window.RenderBackground()
		} else {
			// Set and apply the background color
			renderer.SetDrawColor(
				options.BackgroundColor.R,
				options.BackgroundColor.G,
				options.BackgroundColor.B,
				options.BackgroundColor.A)
			renderer.Clear()
		}

		margins := uniformPadding(30)
		contentWidth := window.Width - (margins.Left + margins.Right)

		var titleHeight int32 = 0
		var titleRect sdl.Rect
		var imageHeight int32 = 0
		var totalContentHeight int32 = 0

		// Render title
		titleFont := fonts.largeFont
		titleSurface, err := titleFont.RenderUTF8Solid(title, options.TitleColor)
		if err == nil {
			defer titleSurface.Free()

			titleRect = sdl.Rect{
				X: (window.Width - titleSurface.W) / 2,
				Y: margins.Top - scrollY,
				W: titleSurface.W,
				H: titleSurface.H,
			}

			titleHeight = titleSurface.H
			totalContentHeight = margins.Top + titleHeight + DefaultTitleSpacing

			if isRectVisible(titleRect, safeAreaHeight) {
				titleTexture, err := renderer.CreateTextureFromSurface(titleSurface)
				if err == nil {
					defer titleTexture.Destroy()
					renderer.Copy(titleTexture, nil, &titleRect)
				}
			}
		}

		imageY := margins.Top + titleHeight + DefaultTitleSpacing - scrollY

		if len(imageTextures) > 0 && currentImageIndex < len(imageTextures) {
			imageRect := imageDimensions[currentImageIndex]
			imageRect.Y = imageY

			if isRectVisible(imageRect, safeAreaHeight) {
				renderer.Copy(imageTextures[currentImageIndex], nil, &imageRect)
			}

			imageHeight = imageRect.H
			totalContentHeight += imageHeight + 30

			if len(imageTextures) > 1 {
				indicatorY := imageY + imageHeight + 10
				indicatorSize := int32(5)
				indicatorSpacing := int32(12)
				totalIndicatorWidth := int32(len(imageTextures))*indicatorSize +
					int32(len(imageTextures)-1)*indicatorSpacing

				indicatorX := (window.Width - totalIndicatorWidth) / 2

				for i := 0; i < len(imageTextures); i++ {
					if i == currentImageIndex {
						renderer.SetDrawColor(255, 255, 255, 255)
					} else {
						renderer.SetDrawColor(100, 100, 100, 255)
					}

					indicatorRect := &sdl.Rect{
						X: indicatorX + int32(i)*(indicatorSize+indicatorSpacing),
						Y: indicatorY,
						W: indicatorSize,
						H: indicatorSize,
					}

					renderer.FillRect(indicatorRect)
				}

				totalContentHeight += indicatorSize + 20
			}
		}

		metadataY := margins.Top + titleHeight + DefaultTitleSpacing
		if imageHeight > 0 {
			metadataY += imageHeight + 30
			if len(imageTextures) > 1 {
				metadataY += 25
			}
		}
		metadataY -= scrollY

		if len(options.Metadata) > 0 {
			metadataFont := fonts.tinyFont
			labelWidth := contentWidth / 4
			valueX := margins.Left + labelWidth
			valueWidth := contentWidth - labelWidth - 10

			currentY := metadataY

			infoLabel := "Info"

			if options.InfoLabel != "" {
				infoLabel = options.InfoLabel
			}

			metadataTitleSurface, err := fonts.smallFont.RenderUTF8Solid(infoLabel, options.TitleColor)
			if err == nil {
				defer metadataTitleSurface.Free()

				metadataTitleRect := &sdl.Rect{
					X: margins.Left,
					Y: currentY,
					W: metadataTitleSurface.W,
					H: metadataTitleSurface.H,
				}

				if isRectVisible(*metadataTitleRect, safeAreaHeight) {
					metadataTitleTexture, err := renderer.CreateTextureFromSurface(metadataTitleSurface)
					if err == nil {
						defer metadataTitleTexture.Destroy()
						renderer.Copy(metadataTitleTexture, nil, metadataTitleRect)
					}
				}

				currentY += metadataTitleSurface.H + 15
			}

			if isLineVisible(margins.Left, currentY, contentWidth, safeAreaHeight) {
				renderer.SetDrawColor(80, 80, 80, 255)
				renderer.DrawLine(
					margins.Left,
					currentY,
					margins.Left+contentWidth,
					currentY)
			}
			currentY += 15

			for _, item := range options.Metadata {
				labelSurface, err := metadataFont.RenderUTF8Solid(item.Label+":", options.MetadataColor)
				if err == nil {
					defer labelSurface.Free()

					labelRect := &sdl.Rect{
						X: margins.Left,
						Y: currentY,
						W: labelSurface.W,
						H: labelSurface.H,
					}

					if isRectVisible(*labelRect, safeAreaHeight) {
						labelTexture, err := renderer.CreateTextureFromSurface(labelSurface)
						if err == nil {
							defer labelTexture.Destroy()
							renderer.Copy(labelTexture, nil, labelRect)
						}
					}
				}

				valueSurface, err := metadataFont.RenderUTF8Blended(item.Value, options.MetadataColor)
				if err == nil {
					defer valueSurface.Free()

					valueRect := &sdl.Rect{
						X: valueX,
						Y: currentY,
						W: min(valueSurface.W, valueWidth),
						H: valueSurface.H,
					}

					if isRectVisible(*valueRect, safeAreaHeight) {
						valueTexture, err := renderer.CreateTextureFromSurface(valueSurface)
						if err == nil {
							defer valueTexture.Destroy()
							renderer.Copy(valueTexture, nil, valueRect)
						}
					}

					currentY += valueSurface.H + 5
				}
			}

			totalContentHeight = currentY + 20 + scrollY
		}

		descriptionY := totalContentHeight - scrollY

		if options.Description != "" {
			descriptionFont := fonts.tinyFont

			descTitleSurface, err := fonts.smallFont.RenderUTF8Solid("Description", options.TitleColor)
			if err == nil {
				defer descTitleSurface.Free()

				descTitleRect := &sdl.Rect{
					X: margins.Left,
					Y: descriptionY,
					W: descTitleSurface.W,
					H: descTitleSurface.H,
				}

				if isRectVisible(*descTitleRect, safeAreaHeight) {
					descTitleTexture, err := renderer.CreateTextureFromSurface(descTitleSurface)
					if err == nil {
						defer descTitleTexture.Destroy()
						renderer.Copy(descTitleTexture, nil, descTitleRect)
					}
				}

				descriptionY += descTitleSurface.H + 15
			}

			if isLineVisible(margins.Left, descriptionY, contentWidth, safeAreaHeight) {
				renderer.SetDrawColor(80, 80, 80, 255)
				renderer.DrawLine(
					margins.Left,
					descriptionY,
					margins.Left+contentWidth,
					descriptionY)
			}
			descriptionY += 15

			renderMultilineText(
				renderer,
				options.Description,
				descriptionFont,
				contentWidth,
				margins.Left,
				descriptionY,
				options.DescriptionColor,
				TextAlignLeft)

			descHeight := calculateMultilineTextHeight(options.Description, descriptionFont, contentWidth)

			totalContentHeight = descriptionY + descHeight + margins.Bottom + scrollY + 35
		}

		// Update maxScrollY based on total content height
		if firstRender {
			maxScrollY = max(0, totalContentHeight-safeAreaHeight+margins.Bottom)
			firstRender = false
		}

		// Render scrollbar with smoother animation
		if options.ShowScrollbar && maxScrollY > 0 {
			scrollbarHeight := int32(float64(safeAreaHeight) * float64(safeAreaHeight) / float64(maxScrollY+safeAreaHeight))
			scrollbarHeight = max(scrollbarHeight, 30) // Minimum height

			scrollbarY := int32(float64(scrollY) * float64(safeAreaHeight-scrollbarHeight) / float64(maxScrollY))

			// Draw scrollbar background
			renderer.SetDrawColor(50, 50, 50, 120)
			renderer.FillRect(&sdl.Rect{
				X: window.Width - scrollbarWidth - 5,
				Y: 5,
				W: scrollbarWidth,
				H: safeAreaHeight - 10,
			})

			// Draw scrollbar handle with rounded corners
			renderer.SetDrawColor(180, 180, 180, 200)
			scrollbarRect := &sdl.Rect{
				X: window.Width - scrollbarWidth - 5,
				Y: 5 + scrollbarY,
				W: scrollbarWidth,
				H: scrollbarHeight,
			}
			drawRoundedRect(renderer, scrollbarRect, 4)
		}

		// Render footer if there are help items
		if len(footerHelpItems) > 0 {
			renderFooter(
				renderer,
				fonts.smallFont,
				footerHelpItems,
				margins.Bottom,
				false,
			)
		}

		renderer.Present()

		// Adjust delay based on whether we're animating
		if abs(scrollY-targetScrollY) > 3 { // Still animating?
			sdl.Delay(8) // Smoother frame rate during animation
		} else {
			sdl.Delay(16) // Standard frame rate when not animating
		}
	}

	// Clean up resources...

	for _, texture := range imageTextures {
		texture.Destroy()
	}

	if result.Cancelled {
		return option.None[DetailScreenReturn](), nil
	}

	return option.Some(result), nil
}

// Helper function to get absolute value of int32
func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

// Helper function to check if a rectangle is visible in the viewport
func isRectVisible(rect sdl.Rect, viewportHeight int32) bool {
	// Check if the rectangle is completely above or below the viewport
	if rect.Y+rect.H < 0 || rect.Y > viewportHeight {
		return false
	}
	return true
}

// Helper function to check if a line is visible in the viewport
func isLineVisible(x, y, width int32, viewportHeight int32) bool {
	// Check if the line is completely above or below the viewport
	if y < 0 || y > viewportHeight {
		return false
	}
	return true
}

func calculateMultilineTextHeight(text string, font *ttf.Font, maxWidth int32) int32 {
	if text == "" {
		return 0
	}

	lines := strings.Split(text, "\n")

	_, fontHeight, err := font.SizeUTF8("Aj")
	if err != nil {
		fontHeight = 20
	}

	lineSpacing := int32(float32(fontHeight) * 0.3)
	totalHeight := int32(0)

	for _, line := range lines {
		if line == "" {
			totalHeight += int32(fontHeight) + lineSpacing
			continue
		}

		remainingText := line
		for len(remainingText) > 0 {
			width, _, err := font.SizeUTF8(remainingText)
			if err != nil || int32(width) <= maxWidth {
				totalHeight += int32(fontHeight) + lineSpacing
				break
			}

			charsPerLine := int(float32(len(remainingText)) * float32(maxWidth) / float32(width))
			if charsPerLine <= 0 {
				charsPerLine = 1
			}

			if charsPerLine < len(remainingText) {

				for i := charsPerLine; i > 0; i-- {
					if i < len(remainingText) && remainingText[i] == ' ' {
						charsPerLine = i
						break
					}
				}
			}

			totalHeight += int32(fontHeight) + lineSpacing
			if charsPerLine >= len(remainingText) {
				break
			}
			remainingText = remainingText[charsPerLine:]
		}
	}

	if totalHeight > lineSpacing {
		totalHeight -= lineSpacing
	}

	totalHeight += 20

	return totalHeight
}
