package gabagool

import (
	"time"

	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type confirmationMessageSettings struct {
	Margins          padding
	MessageText      string
	MessageAlign     TextAlign
	ButtonSpacing    int32
	ConfirmKey       sdl.Keycode
	ConfirmButton    Button
	ImagePath        string
	MaxImageHeight   int32
	MaxImageWidth    int32
	BackgroundColor  sdl.Color
	MessageTextColor sdl.Color
	FooterText       string
	FooterHelpItems  []FooterHelpItem
	FooterTextColor  sdl.Color
	InputDelay       time.Duration
}

type MessageOptions struct {
	ImagePath     string
	ConfirmKey    sdl.Keycode
	ConfirmButton Button
}
type ConfirmationMessageReturn struct {
	Cancelled bool
}

func defaultMessageSettings(message string) confirmationMessageSettings {
	return confirmationMessageSettings{
		Margins:          uniformPadding(20),
		MessageText:      message,
		MessageAlign:     TextAlignCenter,
		ButtonSpacing:    20,
		ConfirmKey:       sdl.K_a,
		ConfirmButton:    ButtonA,
		BackgroundColor:  sdl.Color{R: 0, G: 0, B: 0, A: 255},
		MessageTextColor: sdl.Color{R: 255, G: 255, B: 255, A: 255},
		FooterTextColor:  sdl.Color{R: 180, G: 180, B: 180, A: 255},
		InputDelay:       DefaultInputDelay,
		FooterHelpItems:  []FooterHelpItem{},
	}
}

func ConfirmationMessage(message string, footerHelpItems []FooterHelpItem, options MessageOptions) (types.Option[ConfirmationMessageReturn], error) {
	window := GetWindow()
	renderer := window.Renderer

	settings := defaultMessageSettings(message)
	settings.FooterHelpItems = footerHelpItems

	if options.ImagePath != "" {
		settings.ImagePath = options.ImagePath
		settings.MaxImageHeight = int32(float64(window.Height) / 1.75)
		settings.MaxImageWidth = int32(float64(window.Width) / 1.75)
	}

	if options.ConfirmButton != 0 {
		settings.ConfirmButton = options.ConfirmButton
	}

	if options.ConfirmKey != 0 {
		settings.ConfirmKey = options.ConfirmKey
	}

	result := ConfirmationMessageReturn{Cancelled: true}
	lastInputTime := time.Now()

	imageTexture, imageRect := loadAndPrepareImage(renderer, settings)
	defer func() {
		if imageTexture != nil {
			imageTexture.Destroy()
		}
	}()

	for {
		if !handleEvents(&result, &lastInputTime, settings) {
			break
		}

		renderFrame(renderer, window, settings, imageTexture, imageRect)
		sdl.Delay(16)
	}

	if result.Cancelled {
		return option.None[ConfirmationMessageReturn](), nil
	}
	return option.Some(result), nil
}

func loadAndPrepareImage(renderer *sdl.Renderer, settings confirmationMessageSettings) (*sdl.Texture, sdl.Rect) {
	if settings.ImagePath == "" {
		return nil, sdl.Rect{}
	}

	image, err := img.Load(settings.ImagePath)
	if err != nil {
		return nil, sdl.Rect{}
	}
	defer image.Free()

	imageTexture, err := renderer.CreateTextureFromSurface(image)
	if err != nil {
		return nil, sdl.Rect{}
	}

	widthScale := float32(settings.MaxImageWidth) / float32(image.W)
	heightScale := float32(settings.MaxImageHeight) / float32(image.H)
	scale := widthScale
	if heightScale < widthScale {
		scale = heightScale
	}

	imageW := int32(float32(image.W) * scale)
	imageH := int32(float32(image.H) * scale)

	return imageTexture, sdl.Rect{
		W: imageW,
		H: imageH,
	}
}

func handleEvents(result *ConfirmationMessageReturn, lastInputTime *time.Time, settings confirmationMessageSettings) bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			result.Cancelled = true
			return false

		case *sdl.KeyboardEvent:
			if e.Type != sdl.KEYDOWN || !isInputAllowed(*lastInputTime, settings.InputDelay) {
				continue
			}
			*lastInputTime = time.Now()

			switch e.Keysym.Sym {
			case settings.ConfirmKey, sdl.K_RETURN:
				result.Cancelled = false
				return false
			case sdl.K_b, sdl.K_ESCAPE:
				result.Cancelled = true
				return false
			}

		case *sdl.ControllerButtonEvent:
			if e.Type != sdl.CONTROLLERBUTTONDOWN || !isInputAllowed(*lastInputTime, settings.InputDelay) {
				continue
			}
			*lastInputTime = time.Now()

			switch Button(e.Button) {
			case settings.ConfirmButton:
				result.Cancelled = false
				return false
			case ButtonB:
				result.Cancelled = true
				return false
			}
		}
	}
	return true
}

func isInputAllowed(lastInputTime time.Time, inputDelay time.Duration) bool {
	return time.Since(lastInputTime) >= inputDelay
}

func renderFrame(renderer *sdl.Renderer, window *Window, settings confirmationMessageSettings, imageTexture *sdl.Texture, imageRect sdl.Rect) {
	renderer.SetDrawColor(
		settings.BackgroundColor.R,
		settings.BackgroundColor.G,
		settings.BackgroundColor.B,
		settings.BackgroundColor.A)
	renderer.Clear()

	contentHeight := calculateContentHeight(settings, imageRect)
	startY := (window.Height - contentHeight) / 2

	if imageTexture != nil {
		imageRect.X = (window.Width - imageRect.W) / 2
		imageRect.Y = startY
		renderer.Copy(imageTexture, nil, &imageRect)
		startY = imageRect.Y + imageRect.H + 30
	}

	if len(settings.MessageText) > 0 {
		maxWidth := window.Width - (settings.Margins.Left + settings.Margins.Right)
		renderMultilineText(
			renderer,
			settings.MessageText,
			fonts.smallFont,
			maxWidth,
			window.Width/2,
			startY,
			settings.MessageTextColor)
	}

	renderFooter(
		renderer,
		fonts.smallFont,
		settings.FooterHelpItems,
		settings.Margins.Bottom,
		false,
	)

	renderer.Present()
}

func calculateContentHeight(settings confirmationMessageSettings, imageRect sdl.Rect) int32 {
	var contentHeight int32

	if imageRect.W > 0 {
		contentHeight += imageRect.H + 30
	}

	if len(settings.MessageText) > 0 {
		contentHeight += 30
	}

	return contentHeight
}
