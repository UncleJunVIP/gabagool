package nextui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool/core"
	"github.com/veandco/go-sdl2/sdl"
)

var defaultTheme = core.Theme{
	MainColor:             core.HexToColor(0xFFFFFF),
	PrimaryAccentColor:    core.HexToColor(0x9B2257),
	SecondaryAccentColor:  core.HexToColor(0x1E2329),
	HintInfoColor:         core.HexToColor(0xFFFFFF),
	ListTextColor:         core.HexToColor(0xFFFFFF),
	ListTextSelectedColor: core.HexToColor(0x000000),
	BGColor:               core.HexToColor(0x000000),
}

func InitNextUITheme() core.Theme {
	var nv *NextVal
	var err error

	if core.IsDevMode() {
		nv, err = InitStaticNextVal(os.Getenv("NEXTVAL_PATH"))
	} else {
		nv, err = loadNextVal()
	}

	if err != nil {
		slog.Error("Error loading theme. using default NextUI styling...", "error", err)
		return defaultTheme
	}

	theme := core.Theme{
		MainColor:             parseHexColor(nv.Color1),
		PrimaryAccentColor:    parseHexColor(nv.Color2),
		SecondaryAccentColor:  parseHexColor(nv.Color3),
		ListTextColor:         parseHexColor(nv.Color4),
		ListTextSelectedColor: parseHexColor(nv.Color5),
		HintInfoColor:         parseHexColor(nv.Color6),
		BGColor:               parseHexColor(nv.BGColor),
		FontPath:              nv.FontPath,
	}

	if core.IsDevMode() {
		theme.BackgroundImagePath = os.Getenv(core.BackgroundPathEnvVar)
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

	return core.HexToColor(uint32(hex))
}
