package gabagool

import (
	"strings"
	"time"

	"github.com/veandco/go-sdl2/ttf"

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

type DetailScreenOptions struct {
	Sections            []Section
	TitleColor          sdl.Color
	MetadataColor       sdl.Color
	DescriptionColor    sdl.Color
	BackgroundColor     sdl.Color
	ConfirmButton       Button
	MaxImageHeight      int32
	MaxImageWidth       int32
	ShowScrollbar       bool
	ShowThemeBackground bool
	EnableAction        bool
}

func DefaultInfoScreenOptions() DetailScreenOptions {
	return DetailScreenOptions{
		Sections:         []Section{},
		TitleColor:       sdl.Color{R: 255, G: 255, B: 255, A: 255},
		MetadataColor:    sdl.Color{R: 220, G: 220, B: 220, A: 255},
		DescriptionColor: sdl.Color{R: 200, G: 200, B: 200, A: 255},
		BackgroundColor:  sdl.Color{R: 0, G: 0, B: 0, A: 255},
		ConfirmButton:    ButtonA,
		ShowScrollbar:    true,
		EnableAction:     false,
	}
}

func NewSlideshowSection(title string, imagePaths []string, maxWidth, maxHeight int32) Section {
	return Section{
		Type:       SectionTypeSlideshow,
		Title:      title,
		ImagePaths: imagePaths,
		MaxWidth:   maxWidth,
		MaxHeight:  maxHeight,
	}
}

func NewInfoSection(title string, metadata []MetadataItem) Section {
	return Section{
		Type:     SectionTypeInfo,
		Title:    title,
		Metadata: metadata,
	}
}

func NewDescriptionSection(title string, description string) Section {
	return Section{
		Type:        SectionTypeDescription,
		Title:       title,
		Description: description,
	}
}

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
	LastPressedKey  sdl.Keycode
	LastPressedBtn  uint8
	Cancelled       bool
	ActionTriggered bool
}

func DetailScreen(title string, options DetailScreenOptions, footerHelpItems []FooterHelpItem) (types.Option[DetailScreenReturn], error) {
	window := GetWindow()
	renderer := window.Renderer

	textureCache := newTextureCache()
	defer textureCache.destroy()

	footerHeight := int32(30)
	safeAreaHeight := window.Height - footerHeight

	if options.MaxImageHeight == 0 {
		options.MaxImageHeight = int32(float64(safeAreaHeight) / 2)
	}
	if options.MaxImageWidth == 0 {
		options.MaxImageWidth = int32(float64(window.Width) / 2)
	}

	targetScrollY := int32(0)

	scrollY := int32(0)
	maxScrollY := int32(0)
	scrollSpeed := int32(85)
	scrollbarWidth := int32(10)

	scrollAnimationSpeed := float32(0.15)

	heldDirections := struct {
		up   bool
		down bool
	}{}
	lastRepeatTime := time.Now()
	repeatDelay := time.Millisecond * 100
	repeatInterval := time.Millisecond * 100

	type slideshowState struct {
		currentIndex int
		textures     []*sdl.Texture
		dimensions   []sdl.Rect
	}
	slideshowStates := make(map[int]slideshowState)

	for i, section := range options.Sections {

		if section.Type == SectionTypeSlideshow || section.Type == SectionTypeImage {
			textures := []*sdl.Texture{}
			dimensions := []sdl.Rect{}

			maxWidth := section.MaxWidth
			maxHeight := section.MaxHeight

			if maxWidth == 0 {
				maxWidth = options.MaxImageWidth
			}
			if maxHeight == 0 {
				maxHeight = options.MaxImageHeight
			}

			imagesToLoad := section.ImagePaths
			if section.Type == SectionTypeImage && len(imagesToLoad) > 0 {

				imagesToLoad = imagesToLoad[:1]
			}

			for _, imagePath := range imagesToLoad {
				image, err := img.Load(imagePath)
				if err == nil {

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

						var imageX int32

						if section.Type == SectionTypeImage {

							alignment := TextAlign(section.Alignment)
							switch alignment {
							case AlignLeft:
								imageX = 20
							case AlignRight:
								imageX = window.Width - 20 - imageW
							case AlignCenter:
								fallthrough
							default:
								imageX = (window.Width - imageW) / 2
							}
						} else {

							imageX = (window.Width - imageW) / 2
						}

						rect := sdl.Rect{
							X: imageX,
							Y: 0,
							W: imageW,
							H: imageH,
						}

						dimensions = append(dimensions, rect)
					}
				}
			}

			if len(textures) > 0 {
				slideshowStates[i] = slideshowState{
					currentIndex: 0,
					textures:     textures,
					dimensions:   dimensions,
				}
			}
		}
	}

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

	titleTexture := renderText(renderer, title, fonts.largeFont, options.TitleColor)
	defer func() {
		if titleTexture != nil {
			titleTexture.Destroy()
		}
	}()

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

	handleDirectionalRepeats := func() {
		now := time.Now()
		timeSinceLastRepeat := now.Sub(lastRepeatTime)

		if timeSinceLastRepeat < repeatDelay {
			return
		}

		if repeatInterval > 0 && timeSinceLastRepeat < repeatInterval {
			return
		}

		if heldDirections.up {
			targetScrollY = max32(0, targetScrollY-scrollSpeed)
			lastRepeatTime = now
		} else if heldDirections.down {
			targetScrollY = min32(maxScrollY, targetScrollY+scrollSpeed)
			lastRepeatTime = now
		}
	}

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
						targetScrollY = max32(0, targetScrollY-scrollSpeed)
						lastRepeatTime = currentTime
					case sdl.K_DOWN:
						heldDirections.down = true
						targetScrollY = min32(maxScrollY, targetScrollY+scrollSpeed)
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
					case sdl.K_x:
						if options.EnableAction {
							running = false
							result.Cancelled = false
							result.ActionTriggered = true
						}
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

					switch Button(e.Button) {
					case ButtonUp:
						heldDirections.up = true
						targetScrollY = max32(0, targetScrollY-scrollSpeed)
						lastRepeatTime = currentTime
					case ButtonDown:
						heldDirections.down = true
						targetScrollY = min32(maxScrollY, targetScrollY+scrollSpeed)
						lastRepeatTime = currentTime
					case ButtonLeft:
						if activeSlideshow >= 0 {
							if state, ok := slideshowStates[activeSlideshow]; ok && len(state.textures) > 1 {
								state.currentIndex = (state.currentIndex - 1 + len(state.textures)) % len(state.textures)
								slideshowStates[activeSlideshow] = state
								slideshowIndexChanged = true
							}
						}
					case ButtonRight:
						if activeSlideshow >= 0 {
							if state, ok := slideshowStates[activeSlideshow]; ok && len(state.textures) > 1 {
								state.currentIndex = (state.currentIndex + 1) % len(state.textures)
								slideshowStates[activeSlideshow] = state
								slideshowIndexChanged = true
							}
						}
					case ButtonB:
						result.Cancelled = true
						running = false
					case options.ConfirmButton:
						result.Cancelled = false
						running = false
					case ButtonX:
						if options.EnableAction && e.Type == sdl.CONTROLLERBUTTONDOWN {
							running = false
							result.Cancelled = false
							result.ActionTriggered = true
						}
					}
				} else if e.Type == sdl.CONTROLLERBUTTONUP {
					switch Button(e.Button) {
					case ButtonUp:
						heldDirections.up = false
					case ButtonDown:
						heldDirections.down = false
					}
				}
			}
		}

		handleDirectionalRepeats()

		scrollY += int32(float32(targetScrollY-scrollY) * scrollAnimationSpeed)

		if options.ShowThemeBackground {
			window.RenderBackground()
		} else {

			renderer.SetDrawColor(
				options.BackgroundColor.R,
				options.BackgroundColor.G,
				options.BackgroundColor.B,
				options.BackgroundColor.A)
			renderer.Clear()
		}

		margins := uniformPadding(20)
		contentWidth := window.Width - (margins.Left + margins.Right)

		var titleHeight int32 = 0
		var titleRect sdl.Rect
		var totalContentHeight int32 = 0

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

		for sectionIndex, section := range options.Sections {

			if sectionIndex > 0 {
				currentY += 30
			}

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

			if isLineVisible(margins.Left, currentY, contentWidth, safeAreaHeight) {
				renderer.SetDrawColor(80, 80, 80, 255)
				renderer.DrawLine(
					margins.Left,
					currentY,
					margins.Left+contentWidth,
					currentY)
			}
			currentY += 15

			switch section.Type {
			case SectionTypeSlideshow:
				if state, ok := slideshowStates[sectionIndex]; ok && len(state.textures) > 0 {
					imageRect := state.dimensions[state.currentIndex]
					imageRect.Y = currentY

					if isRectVisible(imageRect, safeAreaHeight) {
						renderer.Copy(state.textures[state.currentIndex], nil, &imageRect)
						activeSlideshow = sectionIndex
					}

					currentY += imageRect.H + 15

					if len(state.textures) > 1 {

						indicatorSize := int32(10)
						indicatorSpacing := int32(5)
						totalIndicatorsWidth := (indicatorSize * int32(len(state.textures))) +
							(indicatorSpacing * int32(len(state.textures)-1))

						indicatorX := (window.Width - totalIndicatorsWidth) / 2
						indicatorY := currentY

						for i := 0; i < len(state.textures); i++ {
							if i == state.currentIndex {
								renderer.SetDrawColor(255, 255, 255, 255)
							} else {
								renderer.SetDrawColor(150, 150, 150, 150)
							}

							indicatorRect := &sdl.Rect{
								X: indicatorX,
								Y: indicatorY,
								W: indicatorSize,
								H: indicatorSize,
							}

							renderer.FillRect(indicatorRect)
							indicatorX += indicatorSize + indicatorSpacing
						}

						currentY += indicatorSize + 15
					}
				}

			case SectionTypeImage:

				if state, ok := slideshowStates[sectionIndex]; ok && len(state.textures) > 0 {
					imageRect := state.dimensions[0]
					imageRect.Y = currentY

					if isRectVisible(imageRect, safeAreaHeight) {
						renderer.Copy(state.textures[0], nil, &imageRect)
					}

					currentY += imageRect.H + 15
				}

			case SectionTypeInfo:
				metadata := section.Metadata

				if len(metadata) > 0 {
					for j, item := range metadata {

						labelTextures, ok := metadataLabelTextures[sectionIndex]
						if !ok || j >= len(labelTextures) || labelTextures[j] == nil {
							continue
						}

						labelTexture := labelTextures[j]

						_, _, labelWidth, labelHeight, _ := labelTexture.Query()
						labelRect := &sdl.Rect{
							X: margins.Left,
							Y: currentY,
							W: labelWidth,
							H: labelHeight,
						}

						if isRectVisible(*labelRect, safeAreaHeight) {
							renderer.Copy(labelTexture, nil, labelRect)
						}

						valueText := item.Value
						if valueText != "" {
							valueX := margins.Left + labelWidth + 10
							maxValueWidth := contentWidth - labelWidth - 10

							valueHeight := calculateMultilineTextHeight(valueText, fonts.smallFont, maxValueWidth)

							if valueHeight > 0 && isRectVisible(sdl.Rect{X: valueX, Y: currentY, W: maxValueWidth, H: valueHeight}, safeAreaHeight) {
								renderMultilineTextOptimized(
									renderer,
									valueText,
									fonts.smallFont,
									maxValueWidth,
									valueX,
									currentY,
									options.MetadataColor,
									AlignLeft,
									textureCache)
							}

							currentY += max32(labelHeight, valueHeight) + 10
						} else {
							currentY += labelHeight + 10
						}
					}

					currentY += 5
				}

			case SectionTypeDescription:

				if section.Description != "" {
					contentWidth := window.Width - (margins.Left + margins.Right)

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

		if firstRender || slideshowIndexChanged {
			maxScrollY = max32(0, totalContentHeight-safeAreaHeight+margins.Bottom)
			if slideshowIndexChanged {
				slideshowIndexChanged = false
			}
			if firstRender {
				firstRender = false
			}
		}

		if options.ShowScrollbar && maxScrollY > 0 {
			scrollbarHeight := int32(float64(safeAreaHeight) * float64(safeAreaHeight) / float64(maxScrollY+safeAreaHeight))
			scrollbarHeight = max32(scrollbarHeight, 30)

			scrollbarY := int32(float64(scrollY) * float64(safeAreaHeight-scrollbarHeight) / float64(maxScrollY))

			renderer.SetDrawColor(50, 50, 50, 120)
			renderer.FillRect(&sdl.Rect{
				X: window.Width - scrollbarWidth - 5,
				Y: 5,
				W: scrollbarWidth,
				H: safeAreaHeight - 10,
			})

			renderer.SetDrawColor(180, 180, 180, 200)
			scrollbarRect := &sdl.Rect{
				X: window.Width - scrollbarWidth - 5,
				Y: 5 + scrollbarY,
				W: scrollbarWidth,
				H: scrollbarHeight,
			}
			drawRoundedRect(renderer, scrollbarRect, 4)
		}

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

		if abs32(scrollY-targetScrollY) > 3 {
			sdl.Delay(8)
		} else {
			sdl.Delay(16)
		}
	}

	if result.Cancelled {
		return option.None[DetailScreenReturn](), nil
	}

	return option.Some(result), nil
}

func renderText(renderer *sdl.Renderer, text string, font *ttf.Font, color sdl.Color) *sdl.Texture {
	if text == "" {
		return nil
	}

	surface, err := font.RenderUTF8Blended(text, color)
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

func renderMultilineTextOptimized(
	renderer *sdl.Renderer,
	text string,
	font *ttf.Font,
	maxWidth int32,
	x, y int32,
	color sdl.Color,
	align TextAlign,
	cache *textureCache) {

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
				lineTexture := cache.get(cacheKey)

				if lineTexture == nil {
					lineSurface, err := font.RenderUTF8Blended(remainingText, color)
					if err == nil {
						lineTexture, err = renderer.CreateTextureFromSurface(lineSurface)
						lineSurface.Free()

						if err == nil {
							cache.set(cacheKey, lineTexture)
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
					default:
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

			lineText := remainingText[:min(charsPerLine, len(remainingText))]
			cacheKey := "line_" + lineText + "_" + string(color.R) + string(color.G) + string(color.B)
			lineTexture := cache.get(cacheKey)

			if lineTexture == nil {
				lineSurface, err := font.RenderUTF8Blended(lineText, color)
				if err == nil {
					lineTexture, err = renderer.CreateTextureFromSurface(lineSurface)
					lineSurface.Free()

					if err == nil {
						cache.set(cacheKey, lineTexture)
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
				default:
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

			remainingText = strings.TrimLeft(remainingText, " ")
		}
	}
}

func isRectVisible(rect sdl.Rect, viewportHeight int32) bool {

	if rect.Y+rect.H < 0 || rect.Y > viewportHeight {
		return false
	}
	return true
}

func isLineVisible(x, y, width int32, viewportHeight int32) bool {

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

			remainingText = strings.TrimLeft(remainingText, " ")
		}
	}

	if totalHeight > lineSpacing {
		totalHeight -= lineSpacing
	}

	totalHeight += 20

	return totalHeight
}
