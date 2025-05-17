package internal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type Theme struct {
	MainColor             sdl.Color
	PrimaryAccentColor    sdl.Color
	SecondaryAccentColor  sdl.Color
	HintInfoColor         sdl.Color
	ListTextColor         sdl.Color
	ListTextSelectedColor sdl.Color
	BGColor               sdl.Color
}

var currentTheme Theme

func initTheme() {
	currentTheme = Theme{
		MainColor:             hexToColor(0xFFFFFF),
		PrimaryAccentColor:    hexToColor(0x9B2257),
		SecondaryAccentColor:  hexToColor(0x1E2329),
		HintInfoColor:         hexToColor(0xFFFFFF),
		ListTextColor:         hexToColor(0xFFFFFF),
		ListTextSelectedColor: hexToColor(0x000000),
		BGColor:               hexToColor(0x000000),
	}

	settingsPath := nextUISettingPath
	if isDev {
		settingsPath = os.Getenv(envSettingsFile)
	}

	file, err := os.Open(settingsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not load theme settings: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "color1", "color2", "color3", "color4", "color5", "color6", "bgcolor":
			color, err := parseHexColor(value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid color value for %s: %s\n", key, value)
				continue
			}

			switch key {
			case "color1":
				currentTheme.MainColor = color
			case "color2":
				currentTheme.PrimaryAccentColor = color
			case "color3":
				currentTheme.SecondaryAccentColor = color
			case "color4":
				currentTheme.ListTextColor = color
			case "color5":
				currentTheme.ListTextSelectedColor = color
			case "color6":
				currentTheme.HintInfoColor = color
			case "bgcolor":
				currentTheme.BGColor = color
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading theme settings: %v\n", err)
		return
	}

}

func parseHexColor(hexStr string) (sdl.Color, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")

	hex, err := strconv.ParseUint(hexStr, 16, 32)
	if err != nil {
		return sdl.Color{}, err
	}

	return hexToColor(uint32(hex)), nil
}

func hexToColor(hex uint32) sdl.Color {
	r := uint8((hex >> 16) & 0xFF)
	g := uint8((hex >> 8) & 0xFF)
	b := uint8(hex & 0xFF)

	return sdl.Color{R: r, G: g, B: b, A: 255}
}

func GetTheme() Theme {
	return currentTheme
}

func GetSDLColorValues(color sdl.Color) (uint8, uint8, uint8, uint8) {
	return color.R, color.G, color.B, color.A
}
