package nextui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/UncleJunVIP/gabagool/v2/pkg/gabagool/constants"
	"github.com/UncleJunVIP/gabagool/v2/pkg/gabagool/internal"
	"github.com/veandco/go-sdl2/sdl"
)

var defaultTheme = internal.Theme{
	MainColor:             internal.HexToColor(0xFFFFFF),
	PrimaryAccentColor:    internal.HexToColor(0x9B2257),
	SecondaryAccentColor:  internal.HexToColor(0x1E2329),
	HintInfoColor:         internal.HexToColor(0xFFFFFF),
	ListTextColor:         internal.HexToColor(0xFFFFFF),
	ListTextSelectedColor: internal.HexToColor(0x000000),
	BGColor:               internal.HexToColor(0x000000),
	FontPath:              "",
	BackgroundImagePath:   "/mnt/SDCARD/bg.png",
}

func InitNextUITheme() internal.Theme {
	var nv *NextVal
	var err error

	if constants.IsDevMode() {
		nv, err = InitStaticNextVal(os.Getenv("NEXTVAL_PATH"))
	} else {
		nv, err = loadNextVal()
	}

	if err != nil {
		return defaultTheme
	}

	theme := internal.Theme{
		MainColor:             parseHexColor(nv.Color1),
		PrimaryAccentColor:    parseHexColor(nv.Color2),
		SecondaryAccentColor:  parseHexColor(nv.Color3),
		ListTextColor:         parseHexColor(nv.Color4),
		ListTextSelectedColor: parseHexColor(nv.Color5),
		HintInfoColor:         parseHexColor(nv.Color6),
		BGColor:               parseHexColor(nv.BGColor),
		FontPath:              nv.FontPath,
	}

	if constants.IsDevMode() {
		theme.BackgroundImagePath = os.Getenv(constants.BackgroundPathEnvVar)
	} else {
		theme.BackgroundImagePath = "/mnt/SDCARD/bg.png"
	}

	return theme
}

func InitStaticNextVal(filePath string) (*NextVal, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var nextval NextVal
	err = json.Unmarshal(data, &nextval)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON from file: %w", err)
	}

	return &nextval, nil
}

func loadNextVal() (*NextVal, error) {
	execPath := "/mnt/SDCARD/.system/tg5040/bin/nextval.elf"

	cmd := exec.Command(execPath)
	output, err := cmd.Output()
	if err != nil {
		internal.GetInternalLogger().Error("Error executing command!", "error", err)
		return nil, err
	}

	jsonStr := strings.TrimSpace(string(output))

	var nextval NextVal
	err = json.Unmarshal([]byte(jsonStr), &nextval)
	if err != nil {
		internal.GetInternalLogger().Error("Error parsing JSON", "error", err)
		return nil, err
	}

	return &nextval, nil
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

	return internal.HexToColor(uint32(hex))
}
