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
	Title            string
	TitleAlign       TextAlignment
	TitleSpacing     int32
	MessageText      string
	MessageAlign     TextAlignment
	ButtonSpacing    int32
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
	ImagePath string
}
type MessageReturn struct {
	ButtonName     string
	LastPressedKey sdl.Keycode
	LastPressedBtn uint8
	Cancelled      bool
}

func defaultMessageSettings(title, message string) messageSettings {
	return messageSettings{
		Margins:          uniformPadding(20),
		Title:            title,
		TitleAlign:       AlignCenter,
		TitleSpacing:     DefaultTitleSpacing,
		MessageText:      message,
		MessageAlign:     AlignCenter,
		ButtonSpacing:    20,
		BackgroundColor:  sdl.Color{R: 0, G: 0, B: 0, A: 255},
		MessageTextColor: sdl.Color{R: 255, G: 255, B: 255, A: 255},
		FooterTextColor:  sdl.Color{R: 180, G: 180, B: 180, A: 255},
		InputDelay:       DefaultInputDelay,
		FooterHelpItems:  []FooterHelpItem{},
	}
}

func Message(title, message string, footerHelpItems []FooterHelpItem, options MessageOptions) (types.Option[MessageReturn], error) {
	window := GetWindow()
	renderer := window.Renderer

	settings := defaultMessageSettings(title, message)
	settings.FooterHelpItems = footerHelpItems

	if options.ImagePath != "" {
		settings.ImagePath = options.ImagePath
		settings.MaxImageHeight = int32(float64(window.Height) / 1.75)
		settings.MaxImageWidth = int32(float64(window.Width) / 1.75)
	}

	running := true
	result := MessageReturn{
		ButtonName:     "",
		LastPressedKey: 0,
		LastPressedBtn: 0,
		Cancelled:      true,
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

				if imageW > settings.MaxImageWidth {
					ratio := float32(settings.MaxImageWidth) / float32(imageW)
					imageW = settings.MaxImageWidth
					imageH = int32(float32(imageH) * ratio)
				}

				if imageH > settings.MaxImageHeight {
					ratio := float32(settings.MaxImageHeight) / float32(imageH)
					imageH = settings.MaxImageHeight
					imageW = int32(float32(imageW) * ratio)
				}
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

				result.LastPressedKey = e.Keysym.Sym

				currentTime := time.Now()
				if currentTime.Sub(lastInputTime) < settings.InputDelay {
					continue
				}
				lastInputTime = currentTime

				switch e.Keysym.Sym {
				case sdl.K_a, sdl.K_RETURN:
					result.ButtonName = "Yes"
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

				result.LastPressedBtn = e.Button

				currentTime := time.Now()
				if currentTime.Sub(lastInputTime) < settings.InputDelay {
					continue
				}
				lastInputTime = currentTime

				switch e.Button {
				case BrickButton_A:
					result.ButtonName = "Yes"
					result.Cancelled = false
					running = false

				case BrickButton_B:
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

		// Calculate initial startY position - will be updated if title renders successfully
		startY := settings.Margins.Top + settings.TitleSpacing + 40

		// Render title
		titleFont := fonts.extraLargeFont
		titleSurface, err := titleFont.RenderUTF8Solid(settings.Title, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if err == nil {
			titleTexture, err := renderer.CreateTextureFromSurface(titleSurface)
			if err == nil {
				titleRect := &sdl.Rect{
					X: (window.Width - titleSurface.W) / 2,
					Y: settings.Margins.Top,
					W: titleSurface.W,
					H: titleSurface.H,
				}
				renderer.Copy(titleTexture, nil, titleRect)
				titleTexture.Destroy()

				// Update startY based on the actual title height
				startY = titleRect.Y + titleRect.H + settings.TitleSpacing
			}
			titleSurface.Free()
		}

		// Render image if available
		if imageTexture != nil {
			imageRect = sdl.Rect{
				X: (window.Width - imageW) / 2,
				Y: startY,
				W: imageW,
				H: imageH,
			}

			renderer.Copy(imageTexture, nil, &imageRect)
			startY = imageRect.Y + imageRect.H + 30 // Update startY after image
		}

		// Render message text
		messageFont := fonts.smallFont
		maxWidth := window.Width - (settings.Margins.Left + settings.Margins.Right)
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
