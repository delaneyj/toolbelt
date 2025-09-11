package toolbelt

import (
    "math"
    "testing"
)

func TestFitEqualsLinearLerp64(t *testing.T) {
    // Compare Fit against direct linear interpolation for several values.
    oldMin, oldMax := 10.0, 30.0
    newMin, newMax := -5.0, 5.0
    for i := 0; i <= 20; i++ {
        x := oldMin + (float64(i)/20.0)*(oldMax-oldMin)
        got := Fit[float64](x, oldMin, oldMax, newMin, newMax)
        want := newMin + ((x-oldMin)*(newMax-newMin))/(oldMax-oldMin)
        if math.Abs(got-want) > 1e-12 {
            t.Fatalf("Fit mismatch at i=%d: got=%v want=%v", i, got, want)
        }
    }
}

func TestFit01UsesLinear32(t *testing.T) {
    newMin, newMax := float32(100), float32(200)
    for i := 0; i <= 20; i++ {
        x := float32(i) / 20
        got := Fit01[float32](x, newMin, newMax)
        want := newMin + x*(newMax-newMin)
        if diff := float64(got - want); math.Abs(diff) > 1e-5 {
            t.Fatalf("Fit01 mismatch at i=%d: got=%v want=%v", i, got, want)
        }
    }
}

func TestEaseLookup(t *testing.T) {
    e := Ease[float64]("easeInOutCubic")
    v := e(0.5)
    // For inOut cubic at t=0.5, value should be 0.5
    if math.Abs(v-0.5) > 1e-12 {
        t.Fatalf("unexpected value from Ease(inoutcubic) at t=0.5: %v", v)
    }

    // Unknown should fall back to linear
    el := Ease[float64]("unknown")
    if el(0.25) != 0.25 {
        t.Fatalf("unknown easing should be linear")
    }
}

