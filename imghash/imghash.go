package imghash

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"math"
	"sync"
)

// --- Color Management & LUTs ---

var (
	linearLUT   [256]float32
	delinearLUT [1024]uint8
)

func init() {
	for i := 0; i < 256; i++ {
		f := float32(i) / 255.0
		if f <= 0.04045 {
			linearLUT[i] = f / 12.92
		} else {
			linearLUT[i] = float32(math.Pow(float64((f+0.055)/1.055), 2.4))
		}
	}
	for i := 0; i < 1024; i++ {
		v := float64(i) / 1023.0
		var res float64
		if v <= 0.0031308 {
			res = 12.92 * v
		} else {
			res = 1.055*math.Pow(v, 1.0/2.4) - 0.055
		}
		delinearLUT[i] = uint8(math.Max(0, math.Min(255, res*255.0)))
	}
}

func rgbToOklab(c color.Color) (l, a, b float32) {
	r1, g1, b1, _ := c.RGBA()
	r, g, bl := linearLUT[uint8(r1>>8)], linearLUT[uint8(g1>>8)], linearLUT[uint8(b1>>8)]

	l0 := 0.4122214708*r + 0.5363325363*g + 0.0514459929*bl
	m0 := 0.2119034982*r + 0.6806995451*g + 0.1073969566*bl
	s0 := 0.0883024619*r + 0.2817188976*g + 0.6299787005*bl

	l_ := float32(math.Cbrt(float64(l0)))
	m_ := float32(math.Cbrt(float64(m0)))
	s_ := float32(math.Cbrt(float64(s0)))

	return 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_,
		1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_,
		0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_
}

// --- DCT Logic ---

type cosineKey struct {
	w  int
	h  int
	nx int
	ny int
}

type cosineTables struct {
	cosX []float32
	cosY []float32
}

var cosineCache sync.Map

func getCosineTables(w, h, nx, ny int) *cosineTables {
	key := cosineKey{w: w, h: h, nx: nx, ny: ny}
	if tables, ok := cosineCache.Load(key); ok {
		return tables.(*cosineTables)
	}
	tables := buildCosineTables(w, h, nx, ny)
	actual, _ := cosineCache.LoadOrStore(key, tables)
	return actual.(*cosineTables)
}

func buildCosineTables(w, h, nx, ny int) *cosineTables {
	cosX := make([]float32, nx*w)
	for x := 0; x < nx; x++ {
		for j := 0; j < w; j++ {
			cosX[x*w+j] = float32(math.Cos(math.Pi * (float64(j) + 0.5) * float64(x) / float64(w)))
		}
	}

	cosY := make([]float32, ny*h)
	for y := 0; y < ny; y++ {
		for i := 0; i < h; i++ {
			cosY[y*h+i] = float32(math.Cos(math.Pi * (float64(i) + 0.5) * float64(y) / float64(h)))
		}
	}

	return &cosineTables{cosX: cosX, cosY: cosY}
}

func applySeparableDCT(data []float32, w, h, nx, ny int, tables *cosineTables) []float32 {
	cosX := tables.cosX
	cosY := tables.cosY
	temp := make([]float32, w*ny)
	for x := 0; x < w; x++ {
		for y := 0; y < ny; y++ {
			var sum float32
			base := y * h
			for i := 0; i < h; i++ {
				sum += data[i*w+x] * cosY[base+i]
			}
			temp[y*w+x] = sum
		}
	}
	coeffs := make([]float32, nx*ny)
	for y := 0; y < ny; y++ {
		for x := 0; x < nx; x++ {
			var sum float32
			base := x * w
			for j := 0; j < w; j++ {
				sum += temp[y*w+j] * cosX[base+j]
			}
			coeffs[y*nx+x] = sum * (2.0 / float32(w*h))
		}
	}
	return coeffs
}

func applyInverseDCT(coeffs []float32, w, h, nx, ny int, tables *cosineTables) []float32 {
	cosX := tables.cosX
	cosY := tables.cosY
	out := make([]float32, w*h)
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			var sum float32
			for y := 0; y < ny; y++ {
				cy := cosY[y*h+i]
				for x := 0; x < nx; x++ {
					c := float32(1.0)
					if x == 0 {
						c *= 0.7071
					}
					if y == 0 {
						c *= 0.7071
					}
					sum += c * coeffs[y*nx+x] * cosX[x*w+j] * cy
				}
			}
			out[i*w+j] = sum
		}
	}
	return out
}

// --- Encoding / Decoding ---

func Encode(img image.Image, nx, ny int) []byte {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	lC, aC, bC := make([]float32, w*h), make([]float32, w*h), make([]float32, w*h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			l, a, bl := rgbToOklab(img.At(b.Min.X+x, b.Min.Y+y))
			lC[y*w+x], aC[y*w+x], bC[y*w+x] = l, a, bl
		}
	}

	tables := getCosineTables(w, h, nx, ny)
	cL := applySeparableDCT(lC, w, h, nx, ny, tables)
	cA := applySeparableDCT(aC, w, h, nx, ny, tables)
	cB := applySeparableDCT(bC, w, h, nx, ny, tables)

	hash := make([]byte, 2+(nx*ny*3))
	hash[0], hash[1] = byte(nx), byte(ny)
	for i, v := range cL {
		hash[2+i] = byte(math.Max(0, math.Min(255, (float64(v)+1.0)*127.5)))
	}
	off := 2 + (nx * ny)
	for i := 0; i < nx*ny; i++ {
		hash[off+i] = byte(math.Max(0, math.Min(255, (float64(cA[i])+1.0)*127.5)))
	}
	off += nx * ny
	for i := 0; i < nx*ny; i++ {
		hash[off+i] = byte(math.Max(0, math.Min(255, (float64(cB[i])+1.0)*127.5)))
	}
	return hash
}

func Decode(hash []byte, w, h int) image.Image {
	nx, ny := int(hash[0]), int(hash[1])
	cL := make([]float32, nx*ny)
	for i := 0; i < nx*ny; i++ {
		cL[i] = (float32(hash[2+i]) / 127.5) - 1.0
	}

	off := 2 + (nx * ny)
	cA, cB := make([]float32, nx*ny), make([]float32, nx*ny)
	for i := 0; i < nx*ny; i++ {
		cA[i] = (float32(hash[off+i]) / 127.5) - 1.0
	}
	off += nx * ny
	for i := 0; i < nx*ny; i++ {
		cB[i] = (float32(hash[off+i]) / 127.5) - 1.0
	}

	tables := getCosineTables(w, h, nx, ny)
	l, a, b := applyInverseDCT(cL, w, h, nx, ny, tables), applyInverseDCT(cA, w, h, nx, ny, tables), applyInverseDCT(cB, w, h, nx, ny, tables)
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < w*h; i++ {
		l_, a_, b_ := float64(l[i]), float64(a[i]), float64(b[i])
		lp, mp, sp := l_+0.396*a_+0.215*b_, l_-0.105*a_-0.063*b_, l_-0.089*a_-1.291*b_
		l0, m0, s0 := lp*lp*lp, mp*mp*mp, sp*sp*sp
		r := +4.076*l0 - 3.307*m0 + 0.230*s0
		g := -1.268*l0 + 2.609*m0 - 0.341*s0
		bl := -0.004*l0 - 0.703*m0 + 1.707*s0
		out.Set(i%w, i/w, color.RGBA{R: fastDelinearize(r), G: fastDelinearize(g), B: fastDelinearize(bl), A: 255})
	}
	return out
}

func fastDelinearize(v float64) uint8 {
	idx := int(v * 1023)
	if idx < 0 {
		return 0
	}
	if idx > 1023 {
		return 255
	}
	return delinearLUT[idx]
}
