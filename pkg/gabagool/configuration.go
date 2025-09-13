package gabagool

import "time"

type Configuration struct {
	PowerButton struct {
		ButtonCode      int           `json:"button_code" yaml:"button_code"`
		DevicePath      string        `json:"device_path" yaml:"device_path"`
		ShortPressMax   time.Duration `json:"short_press_max" yaml:"short_press_max"`
		CoolDownTime    time.Duration `json:"cooldown_time" yaml:"cooldown_time"`
		SuspendScript   string        `json:"suspend_script" yaml:"suspend_script"`
		ShutdownCommand string        `json:"shutdown_command" yaml:"shutdown_command"`
	} `json:"power_button" yaml:"power_button"`

	ButtonMappings struct {
		Up    int `json:"up" yaml:"up"`
		Down  int `json:"down" yaml:"down"`
		Left  int `json:"left" yaml:"left"`
		Right int `json:"right" yaml:"right"`

		A int `json:"a" yaml:"a"`
		B int `json:"b" yaml:"b"`
		X int `json:"x" yaml:"x"`
		Y int `json:"y" yaml:"y"`

		Start  int `json:"start" yaml:"start"`
		Select int `json:"select" yaml:"select"`
		Menu   int `json:"menu" yaml:"menu"`

		L1 int `json:"l1" yaml:"l1"`
		R1 int `json:"r1" yaml:"r1"`

		F1 int `json:"f1" yaml:"f1"`
		F2 int `json:"f2" yaml:"f2"`
	} `json:"button_mappings" yaml:"button_mappings"`

	UI struct {
		Window struct {
			Width  int32 `json:"width" yaml:"width"`
			Height int32 `json:"height" yaml:"height"`
		} `json:"window" yaml:"window"`

		FontSizes struct {
			XLarge int `json:"xlarge" yaml:"xlarge"`
			Large  int `json:"large" yaml:"large"`
			Medium int `json:"medium" yaml:"medium"`
			Small  int `json:"small" yaml:"small"`
			Tiny   int `json:"tiny" yaml:"tiny"`
			Micro  int `json:"micro" yaml:"micro"`
		} `json:"fonts" yaml:"fonts"`

		InputDelay     time.Duration `json:"input_delay" yaml:"input_delay"`
		TitleSpacing   int32         `json:"title_spacing" yaml:"title_spacing"`
		BackgroundPath string        `json:"background_path" yaml:"background_path"`
	} `json:"ui" yaml:"ui"`

	Environment struct {
		SettingsFileEnvVar   string `json:"settings_file_env" yaml:"settings_file_env"`
		BackgroundPathEnvVar string `json:"background_path_env" yaml:"background_path_env"`
		NextUIBackgroundPath string `json:"nextui_background_path" yaml:"nextui_background_path"`
	} `json:"environment" yaml:"environment"`

	Theme struct {
		DefaultFontPath string `json:"default_font_path" yaml:"default_font_path"`
	} `json:"theme" yaml:"theme"`
}

var TempHardcodedConfig = Configuration{
	PowerButton: struct {
		ButtonCode      int           `json:"button_code" yaml:"button_code"`
		DevicePath      string        `json:"device_path" yaml:"device_path"`
		ShortPressMax   time.Duration `json:"short_press_max" yaml:"short_press_max"`
		CoolDownTime    time.Duration `json:"cooldown_time" yaml:"cooldown_time"`
		SuspendScript   string        `json:"suspend_script" yaml:"suspend_script"`
		ShutdownCommand string        `json:"shutdown_command" yaml:"shutdown_command"`
	}{
		ButtonCode:      116,
		DevicePath:      "/dev/input/event1",
		ShortPressMax:   2 * time.Second,
		CoolDownTime:    1 * time.Second,
		SuspendScript:   "/mnt/SDCARD/.system/tg5040/bin/suspend",
		ShutdownCommand: "/sbin/poweroff",
	},

	ButtonMappings: struct {
		Up     int `json:"up" yaml:"up"`
		Down   int `json:"down" yaml:"down"`
		Left   int `json:"left" yaml:"left"`
		Right  int `json:"right" yaml:"right"`
		A      int `json:"a" yaml:"a"`
		B      int `json:"b" yaml:"b"`
		X      int `json:"x" yaml:"x"`
		Y      int `json:"y" yaml:"y"`
		Start  int `json:"start" yaml:"start"`
		Select int `json:"select" yaml:"select"`
		Menu   int `json:"menu" yaml:"menu"`
		L1     int `json:"l1" yaml:"l1"`
		R1     int `json:"r1" yaml:"r1"`
		F1     int `json:"f1" yaml:"f1"`
		F2     int `json:"f2" yaml:"f2"`
	}{
		Up:     11,
		Down:   12,
		Left:   13,
		Right:  14,
		A:      1,
		B:      0,
		X:      3,
		Y:      2,
		Start:  6,
		Select: 4,
		Menu:   5,
		L1:     9,
		R1:     10,
		F1:     7,
		F2:     8,
	},

	UI: struct {
		Window struct {
			Width  int32 `json:"width" yaml:"width"`
			Height int32 `json:"height" yaml:"height"`
		} `json:"window" yaml:"window"`
		FontSizes struct {
			XLarge int `json:"xlarge" yaml:"xlarge"`
			Large  int `json:"large" yaml:"large"`
			Medium int `json:"medium" yaml:"medium"`
			Small  int `json:"small" yaml:"small"`
			Tiny   int `json:"tiny" yaml:"tiny"`
			Micro  int `json:"micro" yaml:"micro"`
		} `json:"fonts" yaml:"fonts"`
		InputDelay     time.Duration `json:"input_delay" yaml:"input_delay"`
		TitleSpacing   int32         `json:"title_spacing" yaml:"title_spacing"`
		BackgroundPath string        `json:"background_path" yaml:"background_path"`
	}{
		Window: struct {
			Width  int32 `json:"width" yaml:"width"`
			Height int32 `json:"height" yaml:"height"`
		}{
			Width:  1024,
			Height: 768,
		},
		FontSizes: struct {
			XLarge int `json:"xlarge" yaml:"xlarge"`
			Large  int `json:"large" yaml:"large"`
			Medium int `json:"medium" yaml:"medium"`
			Small  int `json:"small" yaml:"small"`
			Tiny   int `json:"tiny" yaml:"tiny"`
			Micro  int `json:"micro" yaml:"micro"`
		}{
			XLarge: 22 * 3, // Fix Scaling
			Large:  18 * 3,
			Medium: 16 * 3,
			Small:  12 * 3,
			Tiny:   8 * 3,
			Micro:  6 * 3,
		},
		InputDelay:     20 * time.Millisecond,
		TitleSpacing:   5,
		BackgroundPath: "/mnt/SDCARD/bg.png",
	},

	Environment: struct {
		SettingsFileEnvVar   string `json:"settings_file_env" yaml:"settings_file_env"`
		BackgroundPathEnvVar string `json:"background_path_env" yaml:"background_path_env"`
		NextUIBackgroundPath string `json:"nextui_background_path" yaml:"nextui_background_path"`
	}{
		SettingsFileEnvVar:   "SETTINGS_FILE",
		BackgroundPathEnvVar: "BACKGROUND_PATH",
		NextUIBackgroundPath: "/mnt/SDCARD/bg.png",
	},

	Theme: struct {
		DefaultFontPath string `json:"default_font_path" yaml:"default_font_path"`
	}{
		DefaultFontPath: "/mnt/SDCARD/System/fonts/Cannoli.ttf",
	},
}

func GetConfig() *Configuration {
	return &TempHardcodedConfig
}
