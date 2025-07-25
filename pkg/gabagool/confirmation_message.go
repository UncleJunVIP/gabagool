package gabagool

import (
	"time"

	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type messageSettings struct {
	Margins          padding
	MessageText      string
	MessageAlign     TextAlign
	ButtonSpacing    int32
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
	ConfirmButton Button
}
type MessageReturn struct {
	Cancelled bool
}

func defaultMessageSettings(message string) messageSettings {
	return messageSettings{
		Margins:          uniformPadding(20),
		MessageText:      message,
		MessageAlign:     AlignCenter,
		ButtonSpacing:    20,
		ConfirmButton:    ButtonA,
		BackgroundColor:  sdl.Color{R: 0, G: 0, B: 0, A: 255},
		MessageTextColor: sdl.Color{R: 255, G: 255, B: 255, A: 255},
		FooterTextColor:  sdl.Color{R: 180, G: 180, B: 180, A: 255},
		InputDelay:       DefaultInputDelay,
		FooterHelpItems:  []FooterHelpItem{},
	}
}

func ConfirmationMessage(message string, footerHelpItems []FooterHelpItem, options MessageOptions) (types.Option[MessageReturn], error) {
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

	running := true
	result := MessageReturn{
		Cancelled: true,
	}

	var imageTexture *sdl.Texture
	var imageRect sdl.Rect
	var imageW, imageH int32

	if settings.ImagePath != "" {
		image, err := img.Load(settings.ImagePath)
		if err == nil {
			defer image.Free()

			imageTexture, err = renderer.CreateTextureFromSurface(image)
			if err == nil {
				imageW = image.W
				imageH = image.H

				// Always scale to use the maximum size available
				widthScale := float32(settings.MaxImageWidth) / float32(imageW)
				heightScale := float32(settings.MaxImageHeight) / float32(imageH)

				// Use the smaller scale to maintain aspect ratio while fitting within bounds
				scale := widthScale
				if heightScale < widthScale {
					scale = heightScale
				}

				imageW = int32(float32(imageW) * scale)
				imageH = int32(float32(imageH) * scale)
			}
		}
	}

	lastInputTime := time.Now()

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				result.Cancelled = true

			case *sdl.KeyboardEvent:
				if e.Type != sdl.KEYDOWN {
					continue
				}

				currentTime := time.Now()
				if currentTime.Sub(lastInputTime) < settings.InputDelay {
					continue
				}
				lastInputTime = currentTime

				switch e.Keysym.Sym {
				case sdl.K_a, sdl.K_RETURN:
					result.Cancelled = false
					running = false

				case sdl.K_b, sdl.K_ESCAPE:
					result.Cancelled = true
					running = false
				}

			case *sdl.ControllerButtonEvent:
				if e.Type != sdl.CONTROLLERBUTTONDOWN {
					continue
				}

				currentTime := time.Now()
				if currentTime.Sub(lastInputTime) < settings.InputDelay {
					continue
				}
				lastInputTime = currentTime

				switch Button(e.Button) {
				case settings.ConfirmButton:
					result.Cancelled = false
					running = false

				case ButtonB:
					result.Cancelled = true
					running = false
				}
			}
		}

		renderer.SetDrawColor(
			settings.BackgroundColor.R,
			settings.BackgroundColor.G,
			settings.BackgroundColor.B,
			settings.BackgroundColor.A)
		renderer.Clear()

		var contentHeight int32 = 0

		if imageTexture != nil {
			contentHeight += imageH + 30
		}

		messageFont := fonts.smallFont
		maxWidth := window.Width - (settings.Margins.Left + settings.Margins.Right)
		var messageHeight int32 = 30
		if len(settings.MessageText) > 0 {
			lineCount := (len(settings.MessageText)*8)/int(maxWidth) + 1
			messageHeight = int32(lineCount * 22)
			contentHeight += messageHeight
		}

		startY := (window.Height - contentHeight) / 2

		if imageTexture != nil {
			imageRect = sdl.Rect{
				X: (window.Width - imageW) / 2,
				Y: startY,
				W: imageW,
				H: imageH,
			}

			renderer.Copy(imageTexture, nil, &imageRect)
			startY = imageRect.Y + imageRect.H + 30
		}

		renderMultilineText(
			renderer,
			settings.MessageText,
			messageFont,
			maxWidth,
			window.Width/2,
			startY,
			settings.MessageTextColor)

		renderFooter(
			renderer,
			fonts.smallFont,
			settings.FooterHelpItems,
			settings.Margins.Bottom,
			false,
		)

		renderer.Present()
		sdl.Delay(16)
	}

	if imageTexture != nil {
		imageTexture.Destroy()
	}

	if result.Cancelled {
		return option.None[MessageReturn](), nil
	}

	return option.Some(result), nil
}
