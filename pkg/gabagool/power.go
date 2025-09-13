package gabagool

import (
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/holoplot/go-evdev"
)

// Adapted from https://github.com/ben16w/minui-power-control
func powerButtonHandler(wg *sync.WaitGroup) {
	defer wg.Done()

	config := GetConfig()

	dev, err := evdev.Open(config.PowerButton.DevicePath)
	if err != nil {
		log.Fatalf("Failed to open input device: %v", err)
	}
	log.Printf("Listening on device: %s\n", config.PowerButton.DevicePath)

	var pressTime time.Time
	var cooldownUntil time.Time

	for {
		event, err := dev.ReadOne()
		if err != nil {
			log.Printf("Failed to read input: %v", err)
			continue
		}

		if time.Now().Before(cooldownUntil) {
			continue
		}

		if event.Type == evdev.EV_KEY && event.Code == evdev.EvCode(config.PowerButton.ButtonCode) {
			if event.Value == 0 && !pressTime.IsZero() {
				duration := time.Since(pressTime)
				pressTime = time.Time{}

				if duration < config.PowerButton.ShortPressMax {
					log.Println("Short press detected, suspending...")
					runScript(config.PowerButton.SuspendScript)
					cooldownUntil = time.Now().Add(config.PowerButton.CoolDownTime)
				}
			} else if event.Value == 1 {
				pressTime = time.Now()
			} else if event.Value == 2 {
				duration := time.Since(pressTime)
				if duration >= config.PowerButton.ShortPressMax {
					log.Println("Button held down for 2 seconds, shutting down...")
					runScript(config.PowerButton.ShutdownCommand)
					cooldownUntil = time.Now().Add(config.PowerButton.CoolDownTime)
				}
			}
		}
	}
}

func runScript(scriptPath string) {
	cmd := exec.Command(scriptPath)
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to run %s script: %v", scriptPath, err)
	}
}
