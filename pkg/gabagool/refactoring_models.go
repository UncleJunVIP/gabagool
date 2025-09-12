package gabagool

import "time"

type textScrollData struct {
	needsScrolling      bool
	scrollOffset        int32
	textWidth           int32
	containerWidth      int32
	direction           int
	lastDirectionChange *time.Time
}
