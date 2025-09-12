package gabagool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"os/exec"
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
	FontPath              string
}

type NextVal struct {
	Font            int    `json:"font"`
	Color1          string `json:"color1"`
	Color2          string `json:"color2"`
	Color3          string `json:"color3"`
	Color4          string `json:"color4"`
	Color5          string `json:"color5"`
	Color6          string `json:"color6"`
	BGColor         string `json:"bgcolor"`
	Radius          int    `json:"radius"`
	ShowClock       int    `json:"showclock"`
	Clock24h        int    `json:"clock24h"`
	BatteryPerc     int    `json:"batteryperc"`
	MenuAnim        int    `json:"menuanim"`
	MenuTransitions int    `json:"menutransitions"`
	Recents         int    `json:"recents"`
	GameArt         int    `json:"gameart"`
	ScreenTimeout   int    `json:"screentimeout"`
	SuspendTimeout  int    `json:"suspendTimeout"`
	SwitcherScale   int    `json:"switcherscale"`
	Haptics         int    `json:"haptics"`
	RomFolderBg     int    `json:"romfolderbg"`
	SaveFormat      int    `json:"saveFormat"`
	StateFormat     int    `json:"stateFormat"`
	MuteLeds        int    `json:"muteLeds"`
	ArtWidth        int    `json:"artWidth"`
	Wifi            int    `json:"wifi"`
	FontPath        string `json:"fontpath"`
}

var currentTheme Theme

func initNextUITheme() {
	var nextval *NextVal
	var err error

	if IsDev {
		staticNextVal := os.Getenv(EnvSettingsFile)
		nextval, err = loadStaticNextVal(staticNextVal)
	} else {
		nextval, err = loadNextVal()
	}

	if err != nil {
		slog.Error("Error loading theme... will use default.", "error", err)

		currentTheme = Theme{
			MainColor:             hexToColor(0xFFFFFF),
			PrimaryAccentColor:    hexToColor(0x9B2257),
			SecondaryAccentColor:  hexToColor(0x1E2329),
			HintInfoColor:         hexToColor(0xFFFFFF),
			ListTextColor:         hexToColor(0xFFFFFF),
			ListTextSelectedColor: hexToColor(0x000000),
			BGColor:               hexToColor(0x000000),
		}

		return
	}

	currentTheme.MainColor = parseHexColor(nextval.Color1)
	currentTheme.PrimaryAccentColor = parseHexColor(nextval.Color2)
	currentTheme.SecondaryAccentColor = parseHexColor(nextval.Color3)
	currentTheme.ListTextColor = parseHexColor(nextval.Color4)
	currentTheme.ListTextSelectedColor = parseHexColor(nextval.Color5)
	currentTheme.HintInfoColor = parseHexColor(nextval.Color6)
	currentTheme.BGColor = parseHexColor(nextval.BGColor)
	currentTheme.FontPath = nextval.FontPath
}

func initTheme() {
	currentTheme = Theme{
		MainColor:            hexToColor(0xFFFFFF),
		PrimaryAccentColor:   hexToColor(0x008080),
		SecondaryAccentColor: hexToColor(0x000000),
		HintInfoColor:        hexToColor(0x000000),
		ListTextColor:        hexToColor(0xFFFFFF),

		ListTextSelectedColor: hexToColor(0x000000),
		BGColor:               hexToColor(0xFFFFFF),
		FontPath:              "/mnt/SDCARD/System/fonts/Cannoli.ttf",
	}
}

func parseHexColor(hexStr string) sdl.Color {
	hexStr = strings.TrimPrefix(hexStr, "0x")

	hex, err := strconv.ParseUint(hexStr, 16, 32)
	if err != nil {
		return sdl.Color{
			R: 255,
			G: 0,
			B: 0,
			A: 255,
		}
	}

	return hexToColor(uint32(hex))
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

func loadNextVal() (*NextVal, error) {
	execPath := "/mnt/SDCARD/.system/tg5040/bin/nextval.elf"

	cmd := exec.Command(execPath)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("Error executing command!", "error", err)
		return nil, err
	}

	jsonStr := strings.TrimSpace(string(output))

	var nextval NextVal
	err = json.Unmarshal([]byte(jsonStr), &nextval)
	if err != nil {
		slog.Error("Error parsing JSON", "error", err)
		return nil, err
	}

	return &nextval, nil
}

func loadStaticNextVal(filePath string) (*NextVal, error) {
	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse the JSON
	var nextval NextVal
	err = json.Unmarshal(data, &nextval)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON from file: %w", err)
	}

	return &nextval, nil
}
