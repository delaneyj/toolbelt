package gomps

import (
	"fmt"

	"github.com/zeebo/xxh3"
)

func MiniAvatar(seed string, saturation, lightness uint8, colorCount int, children ...NODE) NODE {
	h := int(xxh3.HashString(seed) & 0x7fffffffffffffff)

	colorCountF := float32(colorCount)
	hue := float32(h%colorCount) * (360 / colorCountF)

	inner := make([]NODE, 25)
	for i := 0; i < 25; i++ {

		if h&(1<<i%15) != 0 {
			x := i / 5
			if i > 14 {
				x = 7 - x
			}
			y := i % 5
			inner[i] = RECT(
				X(float64(x)),
				Y(float64(y)),
				WIDTH(1),
				HEIGHT(1),
			)
		}
	}

	return SVG(
		VIEWBOXF(-1.5, -1.5, 8, 8),
		FILL(fmt.Sprintf("hsl(%f %d%% %d%%)", hue, saturation, lightness)),
		GRP(inner...),
		GRP(children...),
	)
}

func MiniAvatarDefault(seed string, children ...NODE) NODE {
	return MiniAvatar(seed, 95, 45, 9, children...)
}
