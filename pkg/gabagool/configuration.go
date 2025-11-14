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

	UI struct {
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
		SettingsFileEnvVar string `json:"settings_file_env" yaml:"settings_file_env"`
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

	UI: struct {
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
		FontSizes: struct {
			XLarge int `json:"xlarge" yaml:"xlarge"`
			Large  int `json:"large" yaml:"large"`
			Medium int `json:"medium" yaml:"medium"`
			Small  int `json:"small" yaml:"small"`
			Tiny   int `json:"tiny" yaml:"tiny"`
			Micro  int `json:"micro" yaml:"micro"`
		}{
			XLarge: 66,
			Large:  54,
			Medium: 48,
			Small:  36,
			Tiny:   24,
			Micro:  18,
		},
		InputDelay:     20 * time.Millisecond,
		TitleSpacing:   5,
		BackgroundPath: "/mnt/SDCARD/bg.png",
	},

	Environment: struct {
		SettingsFileEnvVar string `json:"settings_file_env" yaml:"settings_file_env"`
	}{
		SettingsFileEnvVar: "SETTINGS_FILE",
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
