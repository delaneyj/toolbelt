package toolbelt

import (
	"math"
)

type EasingMode string

const (
	EasingModeLinear       EasingMode = "linear"
	EasingModeInQuad       EasingMode = "inquad"
	EasingModeOutQuad      EasingMode = "outquad"
	EasingModeInOutQuad    EasingMode = "inoutquad"
	EasingModeInCubic      EasingMode = "incubic"
	EasingModeOutCubic     EasingMode = "outcubic"
	EasingModeInOutCubic   EasingMode = "inoutcubic"
	EasingModeInQuart      EasingMode = "inquart"
	EasingModeOutQuart     EasingMode = "outquart"
	EasingModeInOutQuart   EasingMode = "inoutquart"
	EasingModeInQuint      EasingMode = "inquint"
	EasingModeOutQuint     EasingMode = "outquint"
	EasingModeInOutQuint   EasingMode = "inoutquint"
	EasingModeInSine       EasingMode = "insine"
	EasingModeOutSine      EasingMode = "outsine"
	EasingModeInOutSine    EasingMode = "inoutsine"
	EasingModeInExpo       EasingMode = "inexpo"
	EasingModeOutExpo      EasingMode = "outexpo"
	EasingModeInOutExpo    EasingMode = "inoutexpo"
	EasingModeInCirc       EasingMode = "incirc"
	EasingModeOutCirc      EasingMode = "outcirc"
	EasingModeInOutCirc    EasingMode = "inoutcirc"
	EasingModeInElastic    EasingMode = "inelastic"
	EasingModeOutElastic   EasingMode = "outelastic"
	EasingModeInOutElastic EasingMode = "inoutelastic"
	EasingModeInBack       EasingMode = "inback"
	EasingModeOutBack      EasingMode = "outback"
	EasingModeInOutBack    EasingMode = "inoutback"
	EasingModeInBounce     EasingMode = "inbounce"
	EasingModeOutBounce    EasingMode = "outbounce"
	EasingModeInOutBounce  EasingMode = "inoutbounce"
)

// Easing functions based on Robert Penner's equations.
// t is expected to be in the range [0, 1] and the return value is in [0, 1].

// EasingFunc is a function that takes a normalized time t in [0,1]
// and returns the eased value in [0,1].
type EasingFunc[T Float] func(t T) T

// Linear
func EaseLinear[T Float](t T) T { return t }

// Quad
func EaseInQuad[T Float](t T) T  { return T(float64(t) * float64(t)) }
func EaseOutQuad[T Float](t T) T { return T(float64(t) * (2 - float64(t))) }
func EaseInOutQuad[T Float](t T) T {
	ft := float64(t)
	if ft < 0.5 {
		return T(2 * ft * ft)
	}
	return T(-1 + (4-2*ft)*ft)
}

// Cubic
func EaseInCubic[T Float](t T) T  { ft := float64(t); return T(ft * ft * ft) }
func EaseOutCubic[T Float](t T) T { ft := 1 - float64(t); return T(1 - ft*ft*ft) }
func EaseInOutCubic[T Float](t T) T {
	ft := float64(t)
	if ft < 0.5 {
		return T(4 * ft * ft * ft)
	}
	f := -2*ft + 2
	return T(1 - (f*f*f)/2)
}

// Quart
func EaseInQuart[T Float](t T) T  { ft := float64(t); ft2 := ft * ft; return T(ft2 * ft2) }
func EaseOutQuart[T Float](t T) T { ft := 1 - float64(t); ft2 := ft * ft; return T(1 - ft2*ft2) }
func EaseInOutQuart[T Float](t T) T {
	ft := float64(t)
	if ft < 0.5 {
		return T(8 * ft * ft * ft * ft)
	}
	f := -2*ft + 2
	return T(1 - (f*f*f*f)/2)
}

// Quint
func EaseInQuint[T Float](t T) T  { ft := float64(t); ft2 := ft * ft; return T(ft2 * ft2 * ft) }
func EaseOutQuint[T Float](t T) T { ft := 1 - float64(t); ft2 := ft * ft; return T(1 - ft2*ft2*ft) }
func EaseInOutQuint[T Float](t T) T {
	ft := float64(t)
	if ft < 0.5 {
		return T(16 * ft * ft * ft * ft * ft)
	}
	f := -2*ft + 2
	return T(1 - (f*f*f*f*f)/2)
}

// Sine
func EaseInSine[T Float](t T) T  { return T(1 - math.Cos((float64(t)*math.Pi)/2)) }
func EaseOutSine[T Float](t T) T { return T(math.Sin((float64(t) * math.Pi) / 2)) }
func EaseInOutSine[T Float](t T) T {
	return T(-(math.Cos(math.Pi*float64(t)) - 1) / 2)
}

// Expo
func EaseInExpo[T Float](t T) T {
	ft := float64(t)
	if ft == 0 {
		return 0
	}
	return T(math.Pow(2, 10*ft-10))
}
func EaseOutExpo[T Float](t T) T {
	ft := float64(t)
	if ft == 1 {
		return 1
	}
	return T(1 - math.Pow(2, -10*ft))
}
func EaseInOutExpo[T Float](t T) T {
	ft := float64(t)
	if ft == 0 {
		return 0
	}
	if ft == 1 {
		return 1
	}
	if ft < 0.5 {
		return T(math.Pow(2, 20*ft-10) / 2)
	}
	return T((2 - math.Pow(2, -20*ft+10)) / 2)
}

// Circ
func EaseInCirc[T Float](t T) T  { return T(1 - math.Sqrt(1-math.Pow(float64(t), 2))) }
func EaseOutCirc[T Float](t T) T { ft := float64(t) - 1; return T(math.Sqrt(1 - ft*ft)) }
func EaseInOutCirc[T Float](t T) T {
	ft := float64(t)
	if ft < 0.5 {
		return T((1 - math.Sqrt(1-math.Pow(2*ft, 2))) / 2)
	}
	return T((math.Sqrt(1-math.Pow(-2*ft+2, 2)) + 1) / 2)
}

// Back
func EaseInBack[T Float](t T) T {
	c1 := 1.70158
	c3 := c1 + 1
	ft := float64(t)
	return T(c3*ft*ft*ft - c1*ft*ft)
}
func EaseOutBack[T Float](t T) T {
	c1 := 1.70158
	c3 := c1 + 1
	ft := float64(t) - 1
	return T(1 + c3*ft*ft*ft + c1*ft*ft)
}
func EaseInOutBack[T Float](t T) T {
	c1 := 1.70158
	c2 := c1 * 1.525
	ft := float64(t)
	if ft < 0.5 {
		f := 2 * ft
		return T((f * f * ((c2+1)*f - c2)) / 2)
	}
	f := 2*ft - 2
	return T((f*f*((c2+1)*f+c2) + 2) / 2)
}

// Elastic
func EaseInElastic[T Float](t T) T {
	ft := float64(t)
	if ft == 0 {
		return 0
	}
	if ft == 1 {
		return 1
	}
	c4 := (2 * math.Pi) / 3
	return T(-math.Pow(2, 10*ft-10) * math.Sin((ft*10-10.75)*c4))
}
func EaseOutElastic[T Float](t T) T {
	ft := float64(t)
	if ft == 0 {
		return 0
	}
	if ft == 1 {
		return 1
	}
	c4 := (2 * math.Pi) / 3
	return T(math.Pow(2, -10*ft)*math.Sin((ft*10-0.75)*c4) + 1)
}
func EaseInOutElastic[T Float](t T) T {
	ft := float64(t)
	if ft == 0 {
		return 0
	}
	if ft == 1 {
		return 1
	}
	c5 := (2 * math.Pi) / 4.5
	if ft < 0.5 {
		return T(-(math.Pow(2, 20*ft-10) * math.Sin((20*ft-11.125)*c5)) / 2)
	}
	return T((math.Pow(2, -20*ft+10)*math.Sin((20*ft-11.125)*c5))/2 + 1)
}

// Bounce
func EaseInBounce[T Float](t T) T  { return T(1 - bounceOut(float64(1-t))) }
func EaseOutBounce[T Float](t T) T { return T(bounceOut(float64(t))) }
func EaseInOutBounce[T Float](t T) T {
	ft := float64(t)
	if ft < 0.5 {
		return T((1 - bounceOut(1-2*ft)) / 2)
	}
	return T((1 + bounceOut(2*ft-1)) / 2)
}

// bounceOut is a helper implementing the piecewise bounce easing.
func bounceOut(t float64) float64 {
	n1 := 7.5625
	d1 := 2.75
	if t < 1/d1 {
		return n1 * t * t
	} else if t < 2/d1 {
		t -= 1.5 / d1
		return n1*t*t + 0.75
	} else if t < 2.5/d1 {
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	} else {
		t -= 2.625 / d1
		return n1*t*t + 0.984375
	}
}

// Ease returns a named easing function. Unknown names fall back to linear.
// Accepted names are case-insensitive and allow dashes/underscores/spaces, e.g.:
//
//	"linear", "in-quad", "easeInCubic", "inout-sine", "out_bounce", etc.
func Ease[T Float](mode EasingMode) EasingFunc[T] {

	switch mode {

	// Quad
	case EasingModeInQuad:
		return EaseInQuad[T]
	case EasingModeOutQuad:
		return EaseOutQuad[T]
	case EasingModeInOutQuad:
		return EaseInOutQuad[T]

	// Cubic
	case EasingModeInCubic:
		return EaseInCubic[T]
	case EasingModeOutCubic:
		return EaseOutCubic[T]
	case EasingModeInOutCubic:
		return EaseInOutCubic[T]

	// Quart
	case EasingModeInQuart:
		return EaseInQuart[T]
	case EasingModeOutQuart:
		return EaseOutQuart[T]
	case EasingModeInOutQuart:
		return EaseInOutQuart[T]

	// Quint
	case EasingModeInQuint:
		return EaseInQuint[T]
	case EasingModeOutQuint:
		return EaseOutQuint[T]
	case EasingModeInOutQuint:
		return EaseInOutQuint[T]

	// Sine
	case EasingModeInSine:
		return EaseInSine[T]
	case EasingModeOutSine:
		return EaseOutSine[T]
	case EasingModeInOutSine:
		return EaseInOutSine[T]

	// Expo
	case EasingModeInExpo:
		return EaseInExpo[T]
	case EasingModeOutExpo:
		return EaseOutExpo[T]
	case EasingModeInOutExpo:
		return EaseInOutExpo[T]

	// Circ
	case EasingModeInCirc:
		return EaseInCirc[T]
	case EasingModeOutCirc:
		return EaseOutCirc[T]
	case EasingModeInOutCirc:
		return EaseInOutCirc[T]

	// Back
	case EasingModeInBack:
		return EaseInBack[T]
	case EasingModeOutBack:
		return EaseOutBack[T]
	case EasingModeInOutBack:
		return EaseInOutBack[T]

	// Elastic
	case EasingModeInElastic:
		return EaseInElastic[T]
	case EasingModeOutElastic:
		return EaseOutElastic[T]
	case EasingModeInOutElastic:
		return EaseInOutElastic[T]

	// Bounce
	case EasingModeInBounce:
		return EaseInBounce[T]
	case EasingModeOutBounce:
		return EaseOutBounce[T]
	case EasingModeInOutBounce:
		return EaseInOutBounce[T]

	default:
		return EaseLinear[T]
	}
}
