package ui

import (
	"time"

	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/UncleJunVIP/gabagool/models"
	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type messageSettings struct {
	Margins          models.Padding
	Title            string
	TitleAlign       internal.TextAlignment
	TitleSpacing     int32
	MessageText      string
	MessageAlign     internal.TextAlignment
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

type MessageReturn struct {
	SelectedButton int
	ButtonName     string
	LastPressedKey sdl.Keycode
	LastPressedBtn uint8
	Cancelled      bool
}

func defaultMessageSettings(title, message string) messageSettings {
	return messageSettings{
		Margins:          models.UniformPadding(20),
		Title:            title,
		TitleAlign:       internal.AlignCenter,
		TitleSpacing:     internal.DefaultTitleSpacing,
		MessageText:      message,
		MessageAlign:     internal.AlignCenter,
		ButtonSpacing:    20,
		BackgroundColor:  sdl.Color{R: 0, G: 0, B: 0, A: 255},
		MessageTextColor: sdl.Color{R: 255, G: 255, B: 255, A: 255},
		FooterTextColor:  sdl.Color{R: 180, G: 180, B: 180, A: 255},
		InputDelay:       internal.DefaultInputDelay,
		FooterHelpItems:  []FooterHelpItem{},
	}
}

func Message(title, message string, footerHelpItems []FooterHelpItem, imagePath string) (types.Option[MessageReturn], error) {
	window := internal.GetWindow()
	renderer := window.Renderer

	settings := defaultMessageSettings(title, message)
	settings.FooterHelpItems = footerHelpItems

	if imagePath != "" {
		settings.ImagePath = imagePath
		settings.MaxImageHeight = int32(float64(window.Height) / 1.75)
		settings.MaxImageWidth = int32(float64(window.Width) / 1.75)
	}

	running := true
	result := MessageReturn{
		SelectedButton: -1,
		ButtonName:     "",
		LastPressedKey: 0,
		LastPressedBtn: 0,
		Cancelled:      true,
	}

	var imageTexture *sdl.Texture
	var imageRect sdl.Rect

	if settings.ImagePath != "" {
		image, err := img.Load(settings.ImagePath)
		if err == nil {
			defer image.Free()

			imageTexture, err = renderer.CreateTextureFromSurface(image)
			if err == nil {

				imageW := image.W
				imageH := image.H

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

				imageRect = sdl.Rect{
					X: (window.Width - imageW) / 2,
					Y: settings.Margins.Top + settings.TitleSpacing + 40,
					W: imageW,
					H: imageH,
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
					result.SelectedButton = 0
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
					result.SelectedButton = 0
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

		titleFont := internal.GetXLargeFont()
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
			}
			titleSurface.Free()
		}

		startY := settings.Margins.Top + settings.TitleSpacing + 40

		if imageTexture != nil {
			renderer.Copy(imageTexture, nil, &imageRect)
			startY = imageRect.Y + imageRect.H + 30
		}

		messageFont := internal.GetSmallFont()
		maxWidth := window.Width - (settings.Margins.Left + settings.Margins.Right)
		renderMultilineText(
			renderer,
			settings.MessageText,
			messageFont,
			maxWidth,
			window.Width/2,
			startY,
			settings.MessageTextColor)

		renderMessageFooter(renderer, settings.FooterHelpItems, settings.Margins)

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

func renderMessageFooter(renderer *sdl.Renderer, footerHelpItems []FooterHelpItem, margins models.Padding) {

	if len(footerHelpItems) == 0 {
		return
	}

	_, screenHeight, err := renderer.GetOutputSize()
	if err != nil {
		return
	}

	if len(footerHelpItems) > 0 {
		pillHeight := int32(40)
		pillPadding := int32(12)
		pillSpacing := int32(20)
		font := internal.GetSmallFont()

		whitePillRadius := int32(8)
		blackPillRadius := int32(6)

		buttonPaddingX := int32(12)

		var totalWidth int32 = 0
		var pillInfos []struct {
			buttonName     string
			helpText       string
			buttonWidth    int32
			buttonHeight   int32
			helpWidth      int32
			blackPillWidth int32
			whitePillWidth int32
		}

		for _, item := range footerHelpItems {

			buttonSurface, err := font.RenderUTF8Blended(item.ButtonName, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err != nil {
				continue
			}
			buttonWidth := buttonSurface.W
			buttonHeight := buttonSurface.H
			buttonSurface.Free()

			blackPillWidth := buttonWidth + buttonPaddingX*2

			helpSurface, err := font.RenderUTF8Blended(item.HelpText, sdl.Color{R: 50, G: 50, B: 50, A: 255})
			if err != nil {
				continue
			}
			helpWidth := helpSurface.W
			helpSurface.Free()

			whitePillWidth := blackPillWidth + helpWidth + pillPadding*3

			pillInfos = append(pillInfos, struct {
				buttonName     string
				helpText       string
				buttonWidth    int32
				buttonHeight   int32
				helpWidth      int32
				blackPillWidth int32
				whitePillWidth int32
			}{
				buttonName:     item.ButtonName,
				helpText:       item.HelpText,
				buttonWidth:    buttonWidth,
				buttonHeight:   buttonHeight,
				helpWidth:      helpWidth,
				blackPillWidth: blackPillWidth,
				whitePillWidth: whitePillWidth,
			})

			totalWidth += whitePillWidth + pillSpacing
		}

		screenWidth, _, err := renderer.GetOutputSize()
		if err != nil {
			screenWidth = 768
		}

		startX := (screenWidth - totalWidth + pillSpacing) / 2
		currentX := startX

		for _, info := range pillInfos {

			outerPillRect := &sdl.Rect{
				X: currentX,
				Y: screenHeight - pillHeight - margins.Bottom,
				W: info.whitePillWidth,
				H: pillHeight,
			}

			whiteColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
			drawRoundedRect(renderer, outerPillRect, whitePillRadius, whiteColor)

			blackPillRect := &sdl.Rect{
				X: currentX + pillPadding,
				Y: screenHeight - pillHeight - margins.Bottom + pillPadding/2,
				W: info.blackPillWidth,
				H: pillHeight - pillPadding,
			}
			blackColor := sdl.Color{R: 0, G: 0, B: 0, A: 255}
			drawRoundedRect(renderer, blackPillRect, blackPillRadius, blackColor)

			blackPillCenterX := blackPillRect.X + blackPillRect.W/2
			blackPillCenterY := blackPillRect.Y + blackPillRect.H/2

			buttonSurface, err := font.RenderUTF8Blended(info.buttonName, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err == nil {

				buttonWidth := buttonSurface.W
				buttonHeight := buttonSurface.H

				buttonTexture, err := renderer.CreateTextureFromSurface(buttonSurface)
				if err == nil {

					buttonRect := &sdl.Rect{
						X: blackPillCenterX - buttonWidth/2,
						Y: blackPillCenterY - buttonHeight/2,
						W: buttonWidth,
						H: buttonHeight,
					}
					renderer.Copy(buttonTexture, nil, buttonRect)
					buttonTexture.Destroy()
				}
				buttonSurface.Free()
			}

			helpSurface, err := font.RenderUTF8Blended(info.helpText, sdl.Color{R: 50, G: 50, B: 50, A: 255})
			if err == nil {
				helpTexture, err := renderer.CreateTextureFromSurface(helpSurface)
				if err == nil {
					helpRect := &sdl.Rect{
						X: currentX + info.blackPillWidth + pillPadding*2,
						Y: screenHeight - pillHeight - margins.Bottom + (pillHeight-helpSurface.H)/2,
						W: helpSurface.W,
						H: helpSurface.H,
					}
					renderer.Copy(helpTexture, nil, helpRect)
					helpTexture.Destroy()
				}
				helpSurface.Free()
			}

			currentX += info.whitePillWidth + pillSpacing
		}
	}
}
