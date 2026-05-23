package main

import "github.com/AveryLucas/gogogd/classdb/Label"

// HUD displays the current score as a label. Game calls SetScore
// whenever a coin is collected.
type HUD struct {
	Label.Extension[HUD] `gd:"HUD"`
}

func (h *HUD) Ready() {
	h.SetText("Score: 0")
}

func (h *HUD) SetScore(score int) {
	h.SetText("Score: " + itoa(score))
}

// itoa avoids strconv import for the single use site.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
