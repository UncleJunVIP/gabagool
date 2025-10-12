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
	EnableAction        bool
	ActionKey           sdl.Keycode
	ActionButton        Button
	MaxImageHeight      int32
	MaxImageWidth       int32
	ShowScrollbar       bool
	ShowThemeBackground bool
}

type DetailScreenReturn struct {
	LastPressedKey  sdl.Keycode
	LastPressedBtn  uint8
	Cancelled       bool
	ActionTriggered bool
}

type detailScreenState struct {
	window                *Window
	renderer              *sdl.Renderer
	options               DetailScreenOptions
	footerHelpItems       []FooterHelpItem
	scrollY               int32
	targetScrollY         int32
	maxScrollY            int32
	scrollSpeed           int32
	scrollAnimationSpeed  float32
	lastInputTime         time.Time
	inputDelay            time.Duration
	slideshowStates       map[int]slideshowState
	textureCache          *textureCache
	titleTexture          *sdl.Texture
	sectionTitleTextures  []*sdl.Texture
	metadataLabelTextures map[int][]*sdl.Texture
	heldDirections        struct{ up, down bool }
	lastRepeatTime        time.Time
	repeatDelay           time.Duration
	repeatInterval        time.Duration
	result                DetailScreenReturn
	activeSlideshow       int
}

type slideshowState struct {
	currentIndex int
	textures     []*sdl.Texture
	dimensions   []sdl.Rect
}

func DefaultInfoScreenOptions() DetailScreenOptions {
	return DetailScreenOptions{
		Sections:         []Section{},
		TitleColor:       sdl.Color{R: 255, G: 255, B: 255, A: 255},
		MetadataColor:    sdl.Color{R: 220, G: 220, B: 220, A: 255},
		DescriptionColor: sdl.Color{R: 200, G: 200, B: 200, A: 255},
		BackgroundColor:  sdl.Color{R: 0, G: 0, B: 0, A: 255},
		ActionKey:        sdl.K_a,
		ActionButton:     ButtonA,
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

func DetailScreen(title string, options DetailScreenOptions, footerHelpItems []FooterHelpItem) (types.Option[DetailScreenReturn], error) {
	state := initializeDetailScreenState(title, options, footerHelpItems)
	defer state.cleanup()

	for !state.isFinished() {
		state.handleEvents()
		state.update()
		state.render()
		state.delay()
	}

	// Return None only if explicitly cancelled, otherwise return the result
	if state.result.Cancelled {
		return option.None[DetailScreenReturn](), nil
	}
	return option.Some(state.result), nil
}

func initializeDetailScreenState(title string, options DetailScreenOptions, footerHelpItems []FooterHelpItem) *detailScreenState {
	window := GetWindow()
	state := &detailScreenState{
		window:                window,
		renderer:              window.Renderer,
		options:               options,
		footerHelpItems:       footerHelpItems,
		scrollSpeed:           85,
		scrollAnimationSpeed:  0.15,
		lastInputTime:         time.Now(),
		inputDelay:            DefaultInputDelay,
		slideshowStates:       make(map[int]slideshowState),
		textureCache:          newTextureCache(),
		metadataLabelTextures: make(map[int][]*sdl.Texture),
		repeatDelay:           time.Millisecond * 100,
		repeatInterval:        time.Millisecond * 100,
		result:                DetailScreenReturn{Cancelled: false}, // Start as NOT cancelled
	}

	state.initializeImageDefaults()
	state.loadTextures(title)
	state.initializeSlideshows()

	return state
}

func (s *detailScreenState) initializeImageDefaults() {
	footerHeight := int32(30)
	safeAreaHeight := s.window.Height - footerHeight

	if s.options.MaxImageHeight == 0 {
		s.options.MaxImageHeight = int32(float64(safeAreaHeight) / 2)
	}
	if s.options.MaxImageWidth == 0 {
		s.options.MaxImageWidth = int32(float64(s.window.Width) / 2)
	}
}

func (s *detailScreenState) loadTextures(title string) {
	s.titleTexture = renderText(s.renderer, title, fonts.largeFont, s.options.TitleColor)
	s.sectionTitleTextures = make([]*sdl.Texture, len(s.options.Sections))

	for i, section := range s.options.Sections {
		if section.Title != "" {
			s.sectionTitleTextures[i] = renderText(s.renderer, section.Title, fonts.mediumFont, s.options.TitleColor)
		}

		if section.Type == SectionTypeInfo {
			labelTextures := make([]*sdl.Texture, len(section.Metadata))
			for j, item := range section.Metadata {
				labelTextures[j] = renderText(s.renderer, item.Label+":", fonts.smallFont, s.options.MetadataColor)
			}
			s.metadataLabelTextures[i] = labelTextures
		}
	}
}

func (s *detailScreenState) initializeSlideshows() {
	for i, section := range s.options.Sections {
		if section.Type == SectionTypeSlideshow || section.Type == SectionTypeImage {
			state := s.createSlideshowState(section)
			if len(state.textures) > 0 {
				s.slideshowStates[i] = state
			}
		}
	}
}

func (s *detailScreenState) createSlideshowState(section Section) slideshowState {
	maxWidth := section.MaxWidth
	maxHeight := section.MaxHeight
	if maxWidth == 0 {
		maxWidth = s.options.MaxImageWidth
	}
	if maxHeight == 0 {
		maxHeight = s.options.MaxImageHeight
	}

	imagesToLoad := section.ImagePaths
	if section.Type == SectionTypeImage && len(imagesToLoad) > 0 {
		imagesToLoad = imagesToLoad[:1]
	}

	var textures []*sdl.Texture
	var dimensions []sdl.Rect

	for _, imagePath := range imagesToLoad {
		texture, rect := s.loadAndScaleImage(imagePath, maxWidth, maxHeight, section)
		if texture != nil {
			textures = append(textures, texture)
			dimensions = append(dimensions, rect)
		}
	}

	return slideshowState{
		currentIndex: 0,
		textures:     textures,
		dimensions:   dimensions,
	}
}

func (s *detailScreenState) loadAndScaleImage(imagePath string, maxWidth, maxHeight int32, section Section) (*sdl.Texture, sdl.Rect) {
	image, err := img.Load(imagePath)
	if err != nil {
		return nil, sdl.Rect{}
	}
	defer image.Free()

	imageW, imageH := s.calculateScaledDimensions(image.W, image.H, maxWidth, maxHeight)
	texture, err := s.renderer.CreateTextureFromSurface(image)
	if err != nil {
		return nil, sdl.Rect{}
	}

	imageX := s.calculateImageX(imageW, section)
	return texture, sdl.Rect{X: imageX, Y: 0, W: imageW, H: imageH}
}

func (s *detailScreenState) calculateScaledDimensions(originalW, originalH, maxW, maxH int32) (int32, int32) {
	imageW, imageH := originalW, originalH

	if imageW > maxW {
		ratio := float32(maxW) / float32(imageW)
		imageW = maxW
		imageH = int32(float32(imageH) * ratio)
	}

	if imageH > maxH {
		ratio := float32(maxH) / float32(imageH)
		imageH = maxH
		imageW = int32(float32(imageW) * ratio)
	}

	return imageW, imageH
}

func (s *detailScreenState) calculateImageX(imageW int32, section Section) int32 {
	if section.Type == SectionTypeImage {
		alignment := TextAlign(section.Alignment)
		switch alignment {
		case TextAlignLeft:
			return 20
		case TextAlignRight:
			return s.window.Width - 20 - imageW
		default:
			return (s.window.Width - imageW) / 2
		}
	}
	return (s.window.Width - imageW) / 2
}

func (s *detailScreenState) isFinished() bool {
	// Continue running until we get a definitive exit condition
	return s.result.Cancelled || s.result.ActionTriggered
}

func (s *detailScreenState) handleEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			s.result.Cancelled = true
			return
		case *sdl.KeyboardEvent:
			s.handleKeyboardEvent(e)
		case *sdl.ControllerButtonEvent:
			s.handleControllerEvent(e)
		}
	}
}

func (s *detailScreenState) handleKeyboardEvent(e *sdl.KeyboardEvent) {
	if e.Type == sdl.KEYDOWN {
		s.result.LastPressedKey = e.Keysym.Sym
		if !s.isInputAllowed() {
			return
		}
		s.lastInputTime = time.Now()

		switch e.Keysym.Sym {
		case sdl.K_UP:
			s.startScrolling(true)
		case sdl.K_DOWN:
			s.startScrolling(false)
		case sdl.K_LEFT, sdl.K_RIGHT:
			s.handleSlideshowNavigation(e.Keysym.Sym == sdl.K_LEFT)
		case sdl.K_b, sdl.K_ESCAPE:
			s.result.Cancelled = true
		case sdl.K_a, sdl.K_RETURN:
			s.result.Cancelled = false
		case sdl.K_x:
			if s.options.EnableAction {
				s.result.Cancelled = false
				s.result.ActionTriggered = true
			}
		}
	} else if e.Type == sdl.KEYUP {
		s.stopScrolling(e.Keysym.Sym)
	}
}

func (s *detailScreenState) handleControllerEvent(e *sdl.ControllerButtonEvent) {
	if e.Type == sdl.CONTROLLERBUTTONDOWN {
		s.result.LastPressedBtn = e.Button
		if !s.isInputAllowed() {
			return
		}
		s.lastInputTime = time.Now()

		switch Button(e.Button) {
		case ButtonUp:
			s.startScrolling(true)
		case ButtonDown:
			s.startScrolling(false)
		case ButtonLeft, ButtonRight:
			s.handleSlideshowNavigation(Button(e.Button) == ButtonLeft)
		case ButtonB:
			s.result.Cancelled = true
		case s.options.ActionButton:
			s.result.Cancelled = false
		case ButtonX:
			if s.options.EnableAction {
				s.result.Cancelled = false
				s.result.ActionTriggered = true
			}
		}
	} else if e.Type == sdl.CONTROLLERBUTTONUP {
		s.stopScrollingButton(Button(e.Button))
	}
}

func (s *detailScreenState) isInputAllowed() bool {
	return time.Since(s.lastInputTime) >= s.inputDelay
}

func (s *detailScreenState) startScrolling(up bool) {
	if up {
		s.heldDirections.up = true
		s.targetScrollY = max32(0, s.targetScrollY-s.scrollSpeed)
	} else {
		s.heldDirections.down = true
		s.targetScrollY = min32(s.maxScrollY, s.targetScrollY+s.scrollSpeed)
	}
	s.lastRepeatTime = time.Now()
}

func (s *detailScreenState) stopScrolling(key sdl.Keycode) {
	switch key {
	case sdl.K_UP:
		s.heldDirections.up = false
	case sdl.K_DOWN:
		s.heldDirections.down = false
	}
}

func (s *detailScreenState) stopScrollingButton(button Button) {
	switch button {
	case ButtonUp:
		s.heldDirections.up = false
	case ButtonDown:
		s.heldDirections.down = false
	}
}

func (s *detailScreenState) handleSlideshowNavigation(isLeft bool) {
	activeSlideshow := s.findActiveSlideshow()
	if activeSlideshow >= 0 {
		if state, ok := s.slideshowStates[activeSlideshow]; ok && len(state.textures) > 1 {
			if isLeft {
				state.currentIndex = (state.currentIndex - 1 + len(state.textures)) % len(state.textures)
			} else {
				state.currentIndex = (state.currentIndex + 1) % len(state.textures)
			}
			s.slideshowStates[activeSlideshow] = state
		}
	}
}

func (s *detailScreenState) findActiveSlideshow() int {
	return s.activeSlideshow
}

func (s *detailScreenState) update() {
	s.handleDirectionalRepeats()
	s.scrollY += int32(float32(s.targetScrollY-s.scrollY) * s.scrollAnimationSpeed)
}

func (s *detailScreenState) handleDirectionalRepeats() {
	now := time.Now()
	timeSinceLastRepeat := now.Sub(s.lastRepeatTime)

	if timeSinceLastRepeat < s.repeatDelay {
		return
	}
	if s.repeatInterval > 0 && timeSinceLastRepeat < s.repeatInterval {
		return
	}

	if s.heldDirections.up {
		s.targetScrollY = max32(0, s.targetScrollY-s.scrollSpeed)
		s.lastRepeatTime = now
	} else if s.heldDirections.down {
		s.targetScrollY = min32(s.maxScrollY, s.targetScrollY+s.scrollSpeed)
		s.lastRepeatTime = now
	}
}

func (s *detailScreenState) render() {
	s.clearScreen()

	margins := uniformPadding(20)
	footerHeight := int32(30)
	safeAreaHeight := s.window.Height - footerHeight

	currentY := s.renderTitle(margins)
	currentY, totalContentHeight := s.renderSections(margins, currentY, safeAreaHeight)

	s.updateScrollLimits(totalContentHeight, safeAreaHeight, margins)
	s.renderScrollbar(safeAreaHeight)
	s.renderFooter(margins)

	s.renderer.Present()
}

func (s *detailScreenState) clearScreen() {
	if s.options.ShowThemeBackground {
		s.window.RenderBackground()
	} else {
		s.renderer.SetDrawColor(
			s.options.BackgroundColor.R,
			s.options.BackgroundColor.G,
			s.options.BackgroundColor.B,
			s.options.BackgroundColor.A)
		s.renderer.Clear()
	}
}

func (s *detailScreenState) renderTitle(margins padding) int32 {
	if s.titleTexture == nil {
		return margins.Top + DefaultTitleSpacing - s.scrollY
	}

	_, _, titleW, titleH, err := s.titleTexture.Query()
	if err != nil {
		return margins.Top + DefaultTitleSpacing - s.scrollY
	}

	titleRect := sdl.Rect{
		X: (s.window.Width - titleW) / 2,
		Y: margins.Top - s.scrollY,
		W: titleW,
		H: titleH,
	}

	if isRectVisible(titleRect, s.window.Height) {
		s.renderer.Copy(s.titleTexture, nil, &titleRect)
	}

	return margins.Top + titleH + DefaultTitleSpacing - s.scrollY
}

func (s *detailScreenState) renderSections(margins padding, startY int32, safeAreaHeight int32) (int32, int32) {
	currentY := startY
	contentWidth := s.window.Width - (margins.Left + margins.Right)

	// Reset active slideshow at start of rendering
	s.activeSlideshow = -1

	for sectionIndex, section := range s.options.Sections {
		if sectionIndex > 0 {
			currentY += 30
		}

		currentY = s.renderSectionTitle(sectionIndex, margins, currentY, safeAreaHeight)
		currentY = s.renderSectionDivider(margins, contentWidth, currentY, safeAreaHeight)
		currentY = s.renderSectionContent(sectionIndex, section, margins, contentWidth, currentY, safeAreaHeight)
	}

	return currentY, currentY + s.scrollY + margins.Bottom
}

func (s *detailScreenState) renderSectionTitle(sectionIndex int, margins padding, currentY int32, safeAreaHeight int32) int32 {
	if sectionIndex >= len(s.sectionTitleTextures) || s.sectionTitleTextures[sectionIndex] == nil {
		return currentY
	}

	texture := s.sectionTitleTextures[sectionIndex]
	_, _, titleW, titleH, err := texture.Query()
	if err != nil {
		return currentY
	}

	sectionTitleRect := sdl.Rect{
		X: margins.Left,
		Y: currentY,
		W: titleW,
		H: titleH,
	}

	if isRectVisible(sectionTitleRect, safeAreaHeight) {
		s.renderer.Copy(texture, nil, &sectionTitleRect)
	}

	return currentY + titleH + 15
}

func (s *detailScreenState) renderSectionDivider(margins padding, contentWidth, currentY int32, safeAreaHeight int32) int32 {
	if isLineVisible(currentY, safeAreaHeight) {
		s.renderer.SetDrawColor(80, 80, 80, 255)
		s.renderer.DrawLine(margins.Left, currentY, margins.Left+contentWidth, currentY)
	}
	return currentY + 15
}

func (s *detailScreenState) renderSectionContent(sectionIndex int, section Section, margins padding, contentWidth, currentY int32, safeAreaHeight int32) int32 {
	switch section.Type {
	case SectionTypeSlideshow:
		return s.renderSlideshow(sectionIndex, currentY, safeAreaHeight)
	case SectionTypeImage:
		return s.renderImage(sectionIndex, currentY, safeAreaHeight)
	case SectionTypeInfo:
		return s.renderInfo(sectionIndex, section, margins, contentWidth, currentY, safeAreaHeight)
	case SectionTypeDescription:
		return s.renderDescription(section, margins, contentWidth, currentY, safeAreaHeight)
	}
	return currentY
}

func (s *detailScreenState) renderSlideshow(sectionIndex int, currentY int32, safeAreaHeight int32) int32 {
	state, ok := s.slideshowStates[sectionIndex]
	if !ok || len(state.textures) == 0 {
		return currentY
	}

	imageRect := state.dimensions[state.currentIndex]
	imageRect.Y = currentY

	if isRectVisible(imageRect, safeAreaHeight) {
		s.renderer.Copy(state.textures[state.currentIndex], nil, &imageRect)
		// Set this as the active slideshow when it's being rendered and visible
		s.activeSlideshow = sectionIndex
	}

	currentY += imageRect.H + 15

	if len(state.textures) > 1 {
		currentY = s.renderSlideshowIndicators(state, currentY)
	}

	return currentY
}

func (s *detailScreenState) renderSlideshowIndicators(state slideshowState, currentY int32) int32 {
	indicatorSize := int32(10)
	indicatorSpacing := int32(5)
	totalIndicatorsWidth := (indicatorSize * int32(len(state.textures))) + (indicatorSpacing * int32(len(state.textures)-1))

	indicatorX := (s.window.Width - totalIndicatorsWidth) / 2
	indicatorY := currentY

	for i := 0; i < len(state.textures); i++ {
		if i == state.currentIndex {
			s.renderer.SetDrawColor(255, 255, 255, 255)
		} else {
			s.renderer.SetDrawColor(150, 150, 150, 150)
		}

		indicatorRect := sdl.Rect{
			X: indicatorX,
			Y: indicatorY,
			W: indicatorSize,
			H: indicatorSize,
		}

		s.renderer.FillRect(&indicatorRect)
		indicatorX += indicatorSize + indicatorSpacing
	}

	return currentY + indicatorSize + 15
}

func (s *detailScreenState) renderImage(sectionIndex int, currentY int32, safeAreaHeight int32) int32 {
	state, ok := s.slideshowStates[sectionIndex]
	if !ok || len(state.textures) == 0 {
		return currentY
	}

	imageRect := state.dimensions[0]
	imageRect.Y = currentY

	if isRectVisible(imageRect, safeAreaHeight) {
		s.renderer.Copy(state.textures[0], nil, &imageRect)
	}

	return currentY + imageRect.H + 15
}

func (s *detailScreenState) renderInfo(sectionIndex int, section Section, margins padding, contentWidth, currentY int32, safeAreaHeight int32) int32 {
	labelTextures, ok := s.metadataLabelTextures[sectionIndex]
	if !ok {
		return currentY
	}

	for j, item := range section.Metadata {
		if j >= len(labelTextures) || labelTextures[j] == nil {
			continue
		}

		currentY = s.renderMetadataItem(labelTextures[j], item, margins, contentWidth, currentY, safeAreaHeight)
	}

	return currentY + 5
}

func (s *detailScreenState) renderMetadataItem(labelTexture *sdl.Texture, item MetadataItem, margins padding, contentWidth, currentY int32, safeAreaHeight int32) int32 {
	_, _, labelWidth, labelHeight, _ := labelTexture.Query()
	labelRect := sdl.Rect{
		X: margins.Left,
		Y: currentY,
		W: labelWidth,
		H: labelHeight,
	}

	if isRectVisible(labelRect, safeAreaHeight) {
		s.renderer.Copy(labelTexture, nil, &labelRect)
	}

	if item.Value != "" {
		valueX := margins.Left + labelWidth + 10
		maxValueWidth := contentWidth - labelWidth - 10
		valueHeight := calculateMultilineTextHeight(item.Value, fonts.smallFont, maxValueWidth)

		if valueHeight > 0 && isRectVisible(sdl.Rect{X: valueX, Y: currentY, W: maxValueWidth, H: valueHeight}, safeAreaHeight) {
			renderMultilineTextOptimized(
				s.renderer,
				item.Value,
				fonts.smallFont,
				maxValueWidth,
				valueX,
				currentY,
				s.options.MetadataColor,
				TextAlignLeft,
				s.textureCache)
		}

		return currentY + max32(labelHeight, valueHeight) + 10
	}

	return currentY + labelHeight + 10
}

func (s *detailScreenState) renderDescription(section Section, margins padding, contentWidth, currentY int32, safeAreaHeight int32) int32 {
	if section.Description == "" {
		return currentY
	}

	descHeight := calculateMultilineTextHeight(section.Description, fonts.smallFont, contentWidth)
	if descHeight > 0 && isRectVisible(sdl.Rect{X: margins.Left, Y: currentY, W: contentWidth, H: descHeight}, safeAreaHeight) {
		renderMultilineTextOptimized(
			s.renderer,
			section.Description,
			fonts.smallFont,
			contentWidth,
			margins.Left,
			currentY,
			s.options.DescriptionColor,
			TextAlignLeft,
			s.textureCache)
	}

	return currentY + descHeight + 15
}

func (s *detailScreenState) updateScrollLimits(totalContentHeight int32, safeAreaHeight int32, margins padding) {
	s.maxScrollY = max32(0, totalContentHeight-safeAreaHeight+margins.Bottom)
}

func (s *detailScreenState) renderScrollbar(safeAreaHeight int32) {
	if !s.options.ShowScrollbar || s.maxScrollY <= 0 {
		return
	}

	scrollbarWidth := int32(10)
	scrollbarHeight := int32(float64(safeAreaHeight) * float64(safeAreaHeight) / float64(s.maxScrollY+safeAreaHeight))
	scrollbarHeight = max32(scrollbarHeight, 30)

	scrollbarY := int32(float64(s.scrollY) * float64(safeAreaHeight-scrollbarHeight) / float64(s.maxScrollY))

	// Background
	s.renderer.SetDrawColor(50, 50, 50, 120)
	s.renderer.FillRect(&sdl.Rect{
		X: s.window.Width - scrollbarWidth - 5,
		Y: 5,
		W: scrollbarWidth,
		H: safeAreaHeight - 10,
	})

	// Thumb
	s.renderer.SetDrawColor(180, 180, 180, 200)
	scrollbarRect := &sdl.Rect{
		X: s.window.Width - scrollbarWidth - 5,
		Y: 5 + scrollbarY,
		W: scrollbarWidth,
		H: scrollbarHeight,
	}
	drawRoundedRect(s.renderer, scrollbarRect, 4, sdl.Color{R: 100, G: 100, B: 100, A: 200})
}

func (s *detailScreenState) renderFooter(margins padding) {
	if len(s.footerHelpItems) > 0 {
		renderFooter(
			s.renderer,
			fonts.smallFont,
			s.footerHelpItems,
			margins.Bottom,
			false,
		)
	}
}

func (s *detailScreenState) delay() {
	if abs32(s.scrollY-s.targetScrollY) > 3 {
		sdl.Delay(8)
	} else {
		sdl.Delay(16)
	}
}

func (s *detailScreenState) cleanup() {
	s.textureCache.destroy()

	if s.titleTexture != nil {
		s.titleTexture.Destroy()
	}

	for _, texture := range s.sectionTitleTextures {
		if texture != nil {
			texture.Destroy()
		}
	}

	for _, textures := range s.metadataLabelTextures {
		for _, texture := range textures {
			if texture != nil {
				texture.Destroy()
			}
		}
	}

	for _, state := range s.slideshowStates {
		for _, texture := range state.textures {
			texture.Destroy()
		}
	}
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
					case TextAlignRight:
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
				case TextAlignRight:
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

func isLineVisible(y, viewportHeight int32) bool {
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
