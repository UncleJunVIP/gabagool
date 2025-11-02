package core

import "os"

const Development = "DEV"

const BackgroundPathEnvVar = "BACKGROUND_PATH"

func IsDevMode() bool {
	return os.Getenv("ENVIRONMENT") == Development
}
