package cannoli

import (
	"github.com/UncleJunVIP/gabagool/pkg/gabagool/internal"
)

func InitCannoliTheme(fontPath string) internal.Theme {
	return internal.Theme{
		MainColor:             internal.HexToColor(0xFFFFFF),
		PrimaryAccentColor:    internal.HexToColor(0x008080),
		SecondaryAccentColor:  internal.HexToColor(0x000000),
		HintInfoColor:         internal.HexToColor(0x000000),
		ListTextColor:         internal.HexToColor(0xFFFFFF),
		ListTextSelectedColor: internal.HexToColor(0x000000),
		BGColor:               internal.HexToColor(0xFFFFFF),
		FontPath:              fontPath,
	}

}
