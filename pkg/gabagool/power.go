package gabagool

import (
	"github.com/holoplot/go-evdev"
	"log"
	"os/exec"
	"sync"
	"time"
)

const (
	powerButtonCode = 116
	devicePath      = "/dev/input/event1"
	shortPressMax   = 2 * time.Second
	coolDownTime    = 1 * time.Second
	suspendScript   = "/mnt/SDCARD/.system/tg5040/bin/suspend"
	shutdownCommand = "poweroff -f"
)

// Adapted from https://github.com/ben16w/minui-power-control
func powerButtonHandler(wg *sync.WaitGroup) {
	defer wg.Done()

	dev, err := evdev.Open(devicePath)
	if err != nil {
		log.Fatalf("Failed to open input device: %v", err)
	}
	log.Printf("Listening on device: %s\n", devicePath)

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

		if event.Type == evdev.EV_KEY && event.Code == powerButtonCode {
			if event.Value == 0 && !pressTime.IsZero() {
				duration := time.Since(pressTime)
				pressTime = time.Time{}

				if duration < shortPressMax {
					log.Println("Short press detected, suspending...")
					runScript(suspendScript)
					cooldownUntil = time.Now().Add(coolDownTime)
				}
			} else if event.Value == 1 {
				pressTime = time.Now()
			} else if event.Value == 2 {
				duration := time.Since(pressTime)
				if duration >= shortPressMax {
					log.Println("Button held down for 2 seconds, shutting down...")
					runScript(shutdownCommand)
					cooldownUntil = time.Now().Add(coolDownTime)
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
