package helpers

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func ProgressLoop(msg string, size int, speed int, exit chan struct{}) {
	ticker := time.NewTicker(time.Duration(speed) * time.Millisecond)
	defer ticker.Stop()
	i := 0
	ping := true
	for {
		select {
		case <-exit:
			ticker.Stop()
			return
		case <-ticker.C:
			if i == size {
				ping = false
			} else if i <= 0 {
				ping = true
			}

			nIdx := NormalizeValue(0, float64(size), float64(i))
			mod := int((float64(size) * (-(math.Cos(math.Pi*nIdx) - 1))) / 2)
			fmt.Printf("\r[%s]", strings.Repeat("-", mod)+msg+strings.Repeat("-", size-mod))

			if ping {
				i++
			} else {
				i--
			}
		}

	}
}
