package imghash

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/o1egl/govatar"
)

func TestEncodeDecodeGovatar(t *testing.T) {
	img, err := govatar.Generate(govatar.MALE)
	if err != nil {
		t.Fatalf("generate govatar: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		t.Fatalf("govatar bounds %dx%d, want non-zero", bounds.Dx(), bounds.Dy())
	}

	outDir := govatarOutputDir(t)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}
	writePNG(t, filepath.Join(outDir, "govatar_source.png"), img)

	sizes := parseSizesEnv(t, os.Getenv("IMGHASH_SIZES"))

	for _, size := range sizes {
		t.Run(fmt.Sprintf("%dx%d", size.nx, size.ny), func(t *testing.T) {
			start := time.Now()
			hash := Encode(img, size.nx, size.ny)
			encodeDur := time.Since(start)
			expectedLen := 2 + (size.nx * size.ny * 3)
			if len(hash) != expectedLen {
				t.Fatalf("hash length %d, want %d", len(hash), expectedLen)
			}
			if int(hash[0]) != size.nx || int(hash[1]) != size.ny {
				t.Fatalf("hash header %d,%d, want %d,%d", hash[0], hash[1], size.nx, size.ny)
			}

			start = time.Now()
			decoded := Decode(hash, bounds.Dx(), bounds.Dy())
			decodeDur := time.Since(start)
			decodedBounds := decoded.Bounds()
			if decodedBounds.Dx() != bounds.Dx() || decodedBounds.Dy() != bounds.Dy() {
				t.Fatalf("decoded bounds %dx%d, want %dx%d", decodedBounds.Dx(), decodedBounds.Dy(), bounds.Dx(), bounds.Dy())
			}
			if isMostlyBlank(decoded) {
				t.Fatalf("decoded image appears blank")
			}
			writePNG(t, filepath.Join(outDir, fmt.Sprintf("govatar_decoded_%dx%d.png", size.nx, size.ny)), decoded)
			t.Logf("timing %dx%d encode=%s decode=%s total=%s", size.nx, size.ny, encodeDur, decodeDur, encodeDur+decodeDur)
		})
	}
	t.Logf("wrote images to %s", outDir)
}

func isMostlyBlank(img image.Image) bool {
	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		return true
	}
	points := []image.Point{
		{bounds.Min.X, bounds.Min.Y},
		{bounds.Min.X + bounds.Dx()/2, bounds.Min.Y + bounds.Dy()/2},
		{bounds.Max.X - 1, bounds.Max.Y - 1},
	}
	for _, pt := range points {
		r, g, b, _ := img.At(pt.X, pt.Y).RGBA()
		if r != 0 || g != 0 || b != 0 {
			return false
		}
	}
	return true
}

func govatarOutputDir(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve test file path")
	}
	return filepath.Join(filepath.Dir(filename), "testdata", "govatar_out")
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
}

func parseSizesEnv(t *testing.T, value string) []struct{ nx, ny int } {
	t.Helper()
	if strings.TrimSpace(value) == "" {
		return []struct{ nx, ny int }{
			{4, 4},
			{8, 8},
			{12, 12},
			{16, 16},
		}
	}
	parts := strings.Split(value, ",")
	sizes := make([]struct{ nx, ny int }, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var nx, ny int
		if strings.ContainsAny(part, "xX") {
			fields := strings.FieldsFunc(part, func(r rune) bool { return r == 'x' || r == 'X' })
			if len(fields) != 2 {
				t.Fatalf("invalid IMGHASH_SIZES entry %q", part)
			}
			var err error
			nx, err = strconv.Atoi(strings.TrimSpace(fields[0]))
			if err != nil {
				t.Fatalf("invalid IMGHASH_SIZES nx %q: %v", fields[0], err)
			}
			ny, err = strconv.Atoi(strings.TrimSpace(fields[1]))
			if err != nil {
				t.Fatalf("invalid IMGHASH_SIZES ny %q: %v", fields[1], err)
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil {
				t.Fatalf("invalid IMGHASH_SIZES entry %q: %v", part, err)
			}
			nx, ny = n, n
		}
		if nx <= 0 || ny <= 0 {
			t.Fatalf("invalid IMGHASH_SIZES entry %q: dimensions must be positive", part)
		}
		sizes = append(sizes, struct{ nx, ny int }{nx: nx, ny: ny})
	}
	if len(sizes) == 0 {
		t.Fatalf("IMGHASH_SIZES had no usable entries")
	}
	return sizes
}
