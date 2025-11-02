package cannoli

import (
	"github.com/UncleJunVIP/gabagool/pkg/gabagool/core"
)

func InitCannoliTheme(fontPath string) core.Theme {
	return core.Theme{
		MainColor:             core.HexToColor(0xFFFFFF),
		PrimaryAccentColor:    core.HexToColor(0x008080),
		SecondaryAccentColor:  core.HexToColor(0x000000),
		HintInfoColor:         core.HexToColor(0x000000),
		ListTextColor:         core.HexToColor(0xFFFFFF),
		ListTextSelectedColor: core.HexToColor(0x000000),
		BGColor:               core.HexToColor(0xFFFFFF),
		FontPath:              fontPath,
	}

}
