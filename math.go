package toolbelt

import (
	"math"
	"math/rand"
)

type Float interface {
	~float32 | ~float64
}

type Integer interface {
	~int | ~uint8 | ~int8 | ~uint16 | ~int16 | ~uint32 | ~int32 | ~uint64 | ~int64
}

func Fit[T Float](
	x T,
	oldMin T,
	oldMax T,
	newMin T,
	newMax T,
) T {
	// Normalize x to [0,1], apply linear easing, then scale to new range.
	// This is equivalent to linear interpolation but keeps easing extensible.
	t := (x - oldMin) / (oldMax - oldMin)
	e := EaseLinear(t)
	return newMin + e*(newMax-newMin)
}

func Fit01[T Float](x T, newMin T, newMax T) T {
	// x is expected in [0,1]. Apply linear easing and scale to new range.
	e := EaseLinear(x)
	return newMin + e*(newMax-newMin)
}

func RoundFit01[T Float](x T, newMin T, newMax T) T {
	switch any(x).(type) {
	case float32:
		f := float32(x)
		nmin := float32(newMin)
		nmax := float32(newMax)
		return T(float32(math.Round(float64(Fit01(f, nmin, nmax)))))
	case float64:
		f := float64(x)
		nmin := float64(newMin)
		nmax := float64(newMax)
		return T(math.Round(Fit01(f, nmin, nmax)))
	default:
		panic("unsupported type")
	}
}

func FitMax[T Float](x T, newMax T) T {
	return Fit01(x, 0, newMax)
}

func Clamp[T Float](v T, minimum T, maximum T) T {
	realMin := minimum
	realMax := maximum
	if maximum < realMin {
		realMin = maximum
		realMax = minimum
	}
	return max(realMin, min(realMax, v))
}

func ClampFit[T Float](
	x T,
	oldMin T,
	oldMax T,
	newMin T,
	newMax T,
) T {
	f := Fit(x, oldMin, oldMax, newMin, newMax)
	return Clamp(f, newMin, newMax)
}

func ClampFit01[T Float](x T, newMin T, newMax T) T {
	f := Fit01(x, newMin, newMax)
	return Clamp(f, newMin, newMax)
}

func Clamp01[T Float](v T) T {
	return Clamp(v, 0, 1)
}

func RandNegOneToOneClamped[T Float](r *rand.Rand) T {
	switch any(*new(T)).(type) {
	case float32:
		return T(ClampFit(r.Float32(), 0, 1, -1, 1))
	case float64:
		return T(ClampFit(r.Float64(), 0, 1, -1, 1))
	default:
		panic("unsupported type")
	}
}

func RandIntRange[T Integer](r *rand.Rand, min, max T) T {
	return T(Fit(r.Float32(), 0, 1, float32(min), float32(max)))
}

func RandSliceItem[T any](r *rand.Rand, slice []T) T {
	return slice[r.Intn(len(slice))]
}
