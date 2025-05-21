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

const (
	SectionTypeSlideshow = iota
	SectionTypeInfo
	SectionTypeDescription
	SectionTypeImage
)

// Section struct with all necessary fields
type Section struct {
	Type        int
	Title       string
	ImagePaths  []string
	Metadata    []MetadataItem
	Description string
	MaxWidth    int32
	MaxHeight   int32
	Alignment   int
}

// Updated DetailScreenOptions to use sections
type DetailScreenOptions struct {
	Sections            []Section
	TitleColor          sdl.Color
	MetadataColor       sdl.Color
	DescriptionColor    sdl.Color
	BackgroundColor     sdl.Color
	MaxImageHeight      int32
	MaxImageWidth       int32
	ShowScrollbar       bool
	ShowThemeBackground bool
}

func DefaultInfoScreenOptions() DetailScreenOptions {
	return DetailScreenOptions{
		Sections:         []Section{},
		TitleColor:       sdl.Color{R: 255, G: 255, B: 255, A: 255},
		MetadataColor:    sdl.Color{R: 220, G: 220, B: 220, A: 255},
		DescriptionColor: sdl.Color{R: 200, G: 200, B: 200, A: 255},
		BackgroundColor:  sdl.Color{R: 0, G: 0, B: 0, A: 255},
		ShowScrollbar:    true,
	}
}

// Helper function to create a slideshow section
func NewSlideshowSection(title string, imagePaths []string, maxWidth, maxHeight int32) Section {
	return Section{
		Type:       SectionTypeSlideshow,
		Title:      title,
		ImagePaths: imagePaths,
		MaxWidth:   maxWidth,
		MaxHeight:  maxHeight,
	}
}

// Helper function to create an info section
func NewInfoSection(title string, metadata []MetadataItem) Section {
	return Section{
		Type:     SectionTypeInfo,
		Title:    title,
		Metadata: metadata,
	}
}

// Helper function to create a description section
func NewDescriptionSection(title string, description string) Section {
	return Section{
		Type:        SectionTypeDescription,
		Title:       title,
		Description: description,
	}
}

// Helper function for image section
func NewImageSection(title string, imagePath string, maxWidth, maxHeight int32, alignment TextAlign) Section {
	return Section{
		Type:       SectionTypeImage,
		Title:      title,
		ImagePaths: []string{imagePath},
		MaxWidth:   maxWidth,
		MaxHeight:  maxHeight,
		Alignment:  int(alignment),
	}
}

type DetailScreenReturn struct {
	LastPressedKey sdl.Keycode
	LastPressedBtn uint8
	Cancelled      bool
}

// TextureCache stores pre-rendered textures for reuse
type TextureCache struct {
	textures map[string]*sdl.Texture
}

// NewTextureCache creates a new texture cache
func NewTextureCache() *TextureCache {
	return &TextureCache{
		textures: make(map[string]*sdl.Texture),
	}
}

// Get returns a cached texture if it exists
func (c *TextureCache) Get(key string) *sdl.Texture {
	if texture, exists := c.textures[key]; exists {
		return texture
	}
	return nil
}

// Set adds a texture to the cache
func (c *TextureCache) Set(key string, texture *sdl.Texture) {
	c.textures[key] = texture
}

// Destroy destroys all cached textures and clears the cache
func (c *TextureCache) Destroy() {
	for _, texture := range c.textures {
		texture.Destroy()
	}
	c.textures = make(map[string]*sdl.Texture)
}

func DetailScreen(title string, options DetailScreenOptions, footerHelpItems []FooterHelpItem) (types.Option[DetailScreenReturn], error) {
	window := GetWindow()
	renderer := window.Renderer

	// Create texture cache for text elements
	textureCache := NewTextureCache()
	defer textureCache.Destroy() // Clean up textures when function exits

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

	// Slideshow state for each slideshow section
	type slideshowState struct {
		currentIndex int
		textures     []*sdl.Texture
		dimensions   []sdl.Rect
	}
	slideshowStates := make(map[int]slideshowState)

	// Pre-load all slideshow images
	for i, section := range options.Sections {
		// For both slideshow and image sections, we load images
		if section.Type == SectionTypeSlideshow || section.Type == SectionTypeImage {
			textures := []*sdl.Texture{}
			dimensions := []sdl.Rect{}

			// Get the max dimensions
			maxWidth := section.MaxWidth
			maxHeight := section.MaxHeight

			if maxWidth == 0 {
				maxWidth = options.MaxImageWidth
			}
			if maxHeight == 0 {
				maxHeight = options.MaxImageHeight
			}

			// For slideshows we use all images, for image sections just the first one
			imagesToLoad := section.ImagePaths
			if section.Type == SectionTypeImage && len(imagesToLoad) > 0 {
				// Just to be safe, only use the first image for image sections
				imagesToLoad = imagesToLoad[:1]
			}

			for _, imagePath := range imagesToLoad {
				image, err := img.Load(imagePath)
				if err == nil {
					// Calculate dimensions
					imageW := image.W
					imageH := image.H

					if imageW > maxWidth {
						ratio := float32(maxWidth) / float32(imageW)
						imageW = maxWidth
						imageH = int32(float32(imageH) * ratio)
					}

					if imageH > maxHeight {
						ratio := float32(maxHeight) / float32(imageH)
						imageH = maxHeight
						imageW = int32(float32(imageW) * ratio)
					}

					texture, err := renderer.CreateTextureFromSurface(image)
					image.Free()

					if err == nil {
						textures = append(textures, texture)

						// Calculate X position based on section type and alignment
						var imageX int32

						if section.Type == SectionTypeImage {
							// For image sections, use the alignment
							alignment := TextAlign(section.Alignment)
							switch alignment {
							case AlignLeft:
								imageX = 20 // Use fixed padding for now
							case AlignRight:
								imageX = window.Width - 20 - imageW
							case AlignCenter:
								fallthrough
							default:
								imageX = (window.Width - imageW) / 2
							}
						} else {
							// For slideshows, always center
							imageX = (window.Width - imageW) / 2
						}

						rect := sdl.Rect{
							X: imageX,
							Y: 0, // Y will be set during rendering
							W: imageW,
							H: imageH,
						}

						dimensions = append(dimensions, rect)
					}
				}
			}

			// Only store state if there are textures
			if len(textures) > 0 {
				slideshowStates[i] = slideshowState{
					currentIndex: 0,
					textures:     textures,
					dimensions:   dimensions,
				}
			}
		}
	}

	// Clean up all slideshow textures when function exits
	defer func() {
		for _, state := range slideshowStates {
			for _, texture := range state.textures {
				texture.Destroy()
			}
		}
	}()

	lastInputTime := time.Now()
	inputDelay := DefaultInputDelay

	result := DetailScreenReturn{
		LastPressedKey: 0,
		LastPressedBtn: 0,
		Cancelled:      true,
	}

	running := true
	firstRender := true

	// Pre-render static text elements that won't change
	titleTexture := renderText(renderer, title, fonts.largeFont, options.TitleColor)
	defer func() {
		if titleTexture != nil {
			titleTexture.Destroy()
		}
	}()

	// Pre-render section title textures
	sectionTitleTextures := make([]*sdl.Texture, len(options.Sections))
	for i, section := range options.Sections {
		if section.Title != "" {
			sectionTitleTextures[i] = renderText(renderer, section.Title, fonts.mediumFont, options.TitleColor)
		}
	}
	defer func() {
		for _, texture := range sectionTitleTextures {
			if texture != nil {
				texture.Destroy()
			}
		}
	}()

	// Pre-render metadata label textures for info sections
	metadataLabelTextures := make(map[int][](*sdl.Texture))
	for i, section := range options.Sections {
		if section.Type == SectionTypeInfo {
			labelTextures := make([]*sdl.Texture, len(section.Metadata))

			for j, item := range section.Metadata {
				labelTextures[j] = renderText(renderer, item.Label+":", fonts.smallFont, options.MetadataColor)
			}

			metadataLabelTextures[i] = labelTextures
		}
	}
	defer func() {
		for _, textures := range metadataLabelTextures {
			for _, texture := range textures {
				if texture != nil {
					texture.Destroy()
				}
			}
		}
	}()

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

	// Function to handle slideshow navigation
	slideshowIndexChanged := false
	activeSlideshow := -1

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
						if activeSlideshow >= 0 {
							if state, ok := slideshowStates[activeSlideshow]; ok && len(state.textures) > 1 {
								state.currentIndex = (state.currentIndex - 1 + len(state.textures)) % len(state.textures)
								slideshowStates[activeSlideshow] = state
								slideshowIndexChanged = true
							}
						}
					case sdl.K_RIGHT:
						if activeSlideshow >= 0 {
							if state, ok := slideshowStates[activeSlideshow]; ok && len(state.textures) > 1 {
								state.currentIndex = (state.currentIndex + 1) % len(state.textures)
								slideshowStates[activeSlideshow] = state
								slideshowIndexChanged = true
							}
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
						if activeSlideshow >= 0 {
							if state, ok := slideshowStates[activeSlideshow]; ok && len(state.textures) > 1 {
								state.currentIndex = (state.currentIndex - 1 + len(state.textures)) % len(state.textures)
								slideshowStates[activeSlideshow] = state
								slideshowIndexChanged = true
							}
						}
					case BrickButton_RIGHT:
						if activeSlideshow >= 0 {
							if state, ok := slideshowStates[activeSlideshow]; ok && len(state.textures) > 1 {
								state.currentIndex = (state.currentIndex + 1) % len(state.textures)
								slideshowStates[activeSlideshow] = state
								slideshowIndexChanged = true
							}
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

		// Set background color
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
		var totalContentHeight int32 = 0

		// Render title using pre-rendered texture
		if titleTexture != nil {
			_, _, titleW, titleH, err := titleTexture.Query()
			if err == nil {
				titleRect = sdl.Rect{
					X: (window.Width - titleW) / 2,
					Y: margins.Top - scrollY,
					W: titleW,
					H: titleH,
				}

				titleHeight = titleH
				totalContentHeight = margins.Top + titleHeight + DefaultTitleSpacing

				if isRectVisible(titleRect, safeAreaHeight) {
					renderer.Copy(titleTexture, nil, &titleRect)
				}
			}
		}

		currentY := margins.Top + titleHeight + DefaultTitleSpacing - scrollY
		activeSlideshow = -1

		// Render each section
		for sectionIndex, section := range options.Sections {
			// Add some spacing between sections
			if sectionIndex > 0 {
				currentY += 30
			}

			// Render section title
			sectionTitleTexture := sectionTitleTextures[sectionIndex]
			if sectionTitleTexture != nil {
				_, _, titleW, titleH, err := sectionTitleTexture.Query()
				if err == nil {
					sectionTitleRect := &sdl.Rect{
						X: margins.Left,
						Y: currentY,
						W: titleW,
						H: titleH,
					}

					if isRectVisible(*sectionTitleRect, safeAreaHeight) {
						renderer.Copy(sectionTitleTexture, nil, sectionTitleRect)
					}

					currentY += titleH + 15
				}
			}

			// Render section divider line
			if isLineVisible(margins.Left, currentY, contentWidth, safeAreaHeight) {
				renderer.SetDrawColor(80, 80, 80, 255)
				renderer.DrawLine(
					margins.Left,
					currentY,
					margins.Left+contentWidth,
					currentY)
			}
			currentY += 15

			// Render section content based on type
			// In the rendering loop for sections
			switch section.Type {
			case SectionTypeSlideshow:
				// Keep existing slideshow rendering code
				if state, ok := slideshowStates[sectionIndex]; ok && len(state.textures) > 0 {
					// Existing slideshow code
					imageRect := state.dimensions[state.currentIndex]
					imageRect.Y = currentY

					if isRectVisible(imageRect, safeAreaHeight) {
						renderer.Copy(state.textures[state.currentIndex], nil, &imageRect)
						activeSlideshow = sectionIndex
					}

					currentY += imageRect.H + 15

					// Render indicators only for slideshows with multiple images
					if len(state.textures) > 1 {
						// Existing indicator rendering code
						// ...
					}
				}

			case SectionTypeImage:
				// Image sections are similar but simpler - no indicators or navigation
				if state, ok := slideshowStates[sectionIndex]; ok && len(state.textures) > 0 {
					imageRect := state.dimensions[0] // Always use the first image
					imageRect.Y = currentY

					if isRectVisible(imageRect, safeAreaHeight) {
						renderer.Copy(state.textures[0], nil, &imageRect)
					}

					currentY += imageRect.H + 15
				}

			case SectionTypeInfo:
				// Restore info section rendering
				metadata := section.Metadata

				if len(metadata) > 0 {
					for j, item := range metadata {
						// Get cached label texture
						labelTextures, ok := metadataLabelTextures[sectionIndex]
						if !ok || j >= len(labelTextures) || labelTextures[j] == nil {
							continue
						}

						labelTexture := labelTextures[j]

						// Calculate positions
						_, _, labelWidth, labelHeight, _ := labelTexture.Query()
						labelRect := &sdl.Rect{
							X: margins.Left,
							Y: currentY,
							W: labelWidth,
							H: labelHeight,
						}

						// Render label if visible
						if isRectVisible(*labelRect, safeAreaHeight) {
							renderer.Copy(labelTexture, nil, labelRect)
						}

						// Render value
						valueText := item.Value
						if valueText != "" {
							cacheKey := "value_" + valueText + "_" + section.Title
							valueTexture := textureCache.Get(cacheKey)

							if valueTexture == nil {
								valueTexture = renderText(renderer, valueText, fonts.smallFont, options.MetadataColor)
								if valueTexture != nil {
									textureCache.Set(cacheKey, valueTexture)
								}
							}

							if valueTexture != nil {
								_, _, valueWidth, valueHeight, _ := valueTexture.Query()
								valueRect := &sdl.Rect{
									X: margins.Left + labelWidth + 10,
									Y: currentY,
									W: valueWidth,
									H: valueHeight,
								}

								if isRectVisible(*valueRect, safeAreaHeight) {
									renderer.Copy(valueTexture, nil, valueRect)
								}
							}
						}

						currentY += labelHeight + 10
					}

					currentY += 5
				}

			case SectionTypeDescription:
				// Restore description section rendering
				if section.Description != "" {
					contentWidth := window.Width - (margins.Left + margins.Right)

					// Cache description height calculation
					descHeight := calculateMultilineTextHeight(section.Description, fonts.smallFont, contentWidth)

					if descHeight > 0 && isRectVisible(sdl.Rect{X: margins.Left, Y: currentY, W: contentWidth, H: descHeight}, safeAreaHeight) {
						renderMultilineTextOptimized(
							renderer,
							section.Description,
							fonts.smallFont,
							contentWidth,
							margins.Left,
							currentY,
							options.DescriptionColor,
							AlignLeft,
							textureCache)
					}

					currentY += descHeight + 15
				}
			}
		}

		totalContentHeight = currentY + scrollY + margins.Bottom

		// Update maxScrollY based on total content height
		if firstRender || slideshowIndexChanged {
			maxScrollY = max(0, totalContentHeight-safeAreaHeight+margins.Bottom)
			if slideshowIndexChanged {
				slideshowIndexChanged = false
			}
			if firstRender {
				firstRender = false
			}
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

	if result.Cancelled {
		return option.None[DetailScreenReturn](), nil
	}

	return option.Some(result), nil
}

// Helper function to render text once and return the texture
func renderText(renderer *sdl.Renderer, text string, font *ttf.Font, color sdl.Color) *sdl.Texture {
	if text == "" {
		return nil
	}

	surface, err := font.RenderUTF8Solid(text, color)
	if err != nil {
		return nil
	}
	defer surface.Free()

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil
	}

	return texture
}

// Optimized version of renderMultilineText that uses the texture cache
func renderMultilineTextOptimized(
	renderer *sdl.Renderer,
	text string,
	font *ttf.Font,
	maxWidth int32,
	x, y int32,
	color sdl.Color,
	align TextAlign,
	cache *TextureCache) {

	if text == "" {
		return
	}

	_, fontHeight, err := font.SizeUTF8("Aj")
	if err != nil {
		fontHeight = 20
	}

	lineSpacing := int32(float32(fontHeight) * 0.3)
	lineY := y

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if line == "" {
			lineY += int32(fontHeight) + lineSpacing
			continue
		}

		remainingText := line
		for len(remainingText) > 0 {
			width, _, err := font.SizeUTF8(remainingText)
			if err != nil || int32(width) <= maxWidth {
				cacheKey := "line_" + remainingText + "_" + string(color.R) + string(color.G) + string(color.B)
				lineTexture := cache.Get(cacheKey)

				if lineTexture == nil {
					lineSurface, err := font.RenderUTF8Blended(remainingText, color)
					if err == nil {
						lineTexture, err = renderer.CreateTextureFromSurface(lineSurface)
						lineSurface.Free() // Free surface immediately

						if err == nil {
							cache.Set(cacheKey, lineTexture)
						}
					}
				}

				if lineTexture != nil {
					_, _, lineW, lineH, _ := lineTexture.Query()

					var lineX int32
					switch align {
					case TextAlignCenter:
						lineX = x + (maxWidth-lineW)/2
					case AlignRight:
						lineX = x + maxWidth - lineW
					default: // TextAlignLeft
						lineX = x
					}

					lineRect := &sdl.Rect{
						X: lineX,
						Y: lineY,
						W: lineW,
						H: lineH,
					}

					renderer.Copy(lineTexture, nil, lineRect)
				}

				lineY += int32(fontHeight) + lineSpacing
				break
			}

			// If text is too long, find a good breaking point
			charsPerLine := int(float32(len(remainingText)) * float32(maxWidth) / float32(width))
			if charsPerLine <= 0 {
				charsPerLine = 1
			}

			if charsPerLine < len(remainingText) {
				// Look for a space to break at
				for i := charsPerLine; i > 0; i-- {
					if i < len(remainingText) && remainingText[i] == ' ' {
						charsPerLine = i
						break
					}
				}
			}

			lineText := remainingText[:min(charsPerLine, len(remainingText))]
			cacheKey := "line_" + lineText + "_" + string(color.R) + string(color.G) + string(color.B)
			lineTexture := cache.Get(cacheKey)

			if lineTexture == nil {
				lineSurface, err := font.RenderUTF8Blended(lineText, color)
				if err == nil {
					lineTexture, err = renderer.CreateTextureFromSurface(lineSurface)
					lineSurface.Free() // Free surface immediately

					if err == nil {
						cache.Set(cacheKey, lineTexture)
					}
				}
			}

			if lineTexture != nil {
				_, _, lineW, lineH, _ := lineTexture.Query()

				var lineX int32
				switch align {
				case TextAlignCenter:
					lineX = x + (maxWidth-lineW)/2
				case AlignRight:
					lineX = x + maxWidth - lineW
				default: // TextAlignLeft
					lineX = x
				}

				lineRect := &sdl.Rect{
					X: lineX,
					Y: lineY,
					W: lineW,
					H: lineH,
				}

				renderer.Copy(lineTexture, nil, lineRect)
			}

			lineY += int32(fontHeight) + lineSpacing

			if charsPerLine >= len(remainingText) {
				break
			}

			remainingText = remainingText[charsPerLine:]
			// Skip leading spaces in the next line
			remainingText = strings.TrimLeft(remainingText, " ")
		}
	}
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
			// Skip leading spaces in the next line
			remainingText = strings.TrimLeft(remainingText, " ")
		}
	}

	if totalHeight > lineSpacing {
		totalHeight -= lineSpacing
	}

	totalHeight += 20

	return totalHeight
}
