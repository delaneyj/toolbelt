package gomps

import (
	"fmt"

	"github.com/delaneyj/toolbelt"
	"github.com/zeebo/xxh3"
)

func MiniAvatar(seed string, saturation, lightness uint8, colorCount, gridSize int, children ...NODE) NODE {
	cellCount := gridSize * gridSize
	h := int(xxh3.HashString(seed))

	colorCountF := float32(colorCount)
	hue := toolbelt.Fit(float32(h%colorCount), 0, colorCountF, 0, 360)

	half := 0.5
	inner := make([]NODE, 0, cellCount)
	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			o := y*gridSize + x

			// check if o bit is set
			if (h>>uint(o))&1 == 1 {
				inner = append(inner, CIRCLE(
					CX(float64(x)+half),
					CY(float64(y)+half),
					R(half),
					// WIDTH(1),
					// HEIGHT(1),
				))
			}
		}
	}

	hsl := fmt.Sprintf("hsl(%f %d%% %d%%)", hue, saturation, lightness)

	return SVG(
		VIEWBOXF(0, 0, float64(gridSize), float64(gridSize)),
		FILL(hsl),
		// STROKE(hsl),
		// STROKEWIDTH(0.1),
		GRP(inner...),
		GRP(children...),
	)
}

func MiniAvatarDefault(seed string, children ...NODE) NODE {
	return MiniAvatar(seed, 95, 45, 8, 5, children...)
}
