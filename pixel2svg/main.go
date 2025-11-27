package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

type CLI struct {
	Input      string `arg:"" required:"" help:"Input PNG/GIF file"`
	Output     string `arg:"" optional:"" help:"Output SVG file (defaults to input name with .svg extension)"`
	Scale      int    `short:"s" default:"1" help:"Scale factor for output"`
	Verbose    bool   `short:"v" help:"Show optimization details"`
	Animated   bool   `short:"a" help:"Generate animated SVG for GIF files"`
	FrameFiles bool   `short:"f" help:"Output separate SVG files for each frame"`
}

func run(ctx context.Context) error {
	godotenv.Load()

	var cli CLI
	kong.Parse(&cli)

	return convert(ctx, &cli)
}

type pixelRun struct {
	x, y, width, height int
	color               color.RGBA
}

type optimizationConfig struct {
	name        string
	cssColors   bool
	groupColors bool
	usePaths    bool
	pack2D      bool
	pixelTrace  bool
}

func convertAnimatedGIF(ctx context.Context, cmd *CLI, g *gif.GIF) error {
	if cmd.FrameFiles {
		// Output separate SVG files for each frame
		return convertGIFToFrameFiles(ctx, cmd, g)
	} else {
		// Generate a single animated SVG
		return convertGIFToAnimatedSVG(ctx, cmd, g)
	}
}

func convertGIFToFrameFiles(ctx context.Context, cmd *CLI, g *gif.GIF) error {
	// Get base filename without extension
	base := strings.TrimSuffix(cmd.Output, filepath.Ext(cmd.Output))

	for i, frame := range g.Image {
		// Create filename for this frame
		frameOutput := fmt.Sprintf("%s_frame%03d.svg", base, i)

		// Update command with frame-specific output
		frameCmd := *cmd
		frameCmd.Output = frameOutput
		frameCmd.Verbose = false // Don't repeat verbose output for each frame

		log.Printf("Converting frame %d/%d...", i+1, len(g.Image))
		if err := convertSingleImage(ctx, &frameCmd, frame); err != nil {
			return fmt.Errorf("failed to convert frame %d: %w", i, err)
		}
	}

	log.Printf("Generated %d SVG files: %s_frame*.svg", len(g.Image), base)
	return nil
}

func convertGIFToAnimatedSVG(ctx context.Context, cmd *CLI, g *gif.GIF) error {
	// For animated SVG, we need to process all frames and combine them
	// First, let's find the optimal configuration by testing on the first frame
	bounds := g.Image[0].Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, g.Image[0], bounds.Min, draw.Src)

	configs := generateOptimizationConfigs()
	bestConfig, _, _ := findBestOptimization(rgba, bounds, cmd, configs)

	// Create output file
	out, err := os.Create(cmd.Output)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Generate animated SVG
	generateAnimatedSVG(out, g, bounds, cmd, bestConfig)

	log.Printf("Generated animated SVG: %s (%d frames)", cmd.Output, len(g.Image))
	return nil
}

func generateAnimatedSVG(w io.Writer, g *gif.GIF, bounds image.Rectangle, cmd *CLI, config optimizationConfig) {
	baseWidth := bounds.Dx()
	baseHeight := bounds.Dy()
	scaledWidth := baseWidth * cmd.Scale
	scaledHeight := baseHeight * cmd.Scale

	// Write SVG header
	fmt.Fprintf(w, `<?xml version="1.0"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`,
		scaledWidth, scaledHeight, baseWidth, baseHeight)

	// Calculate total duration
	totalDuration := 0
	for _, delay := range g.Delay {
		totalDuration += delay * 10 // Delay is in 100ths of a second, convert to ms
	}
	if totalDuration == 0 {
		totalDuration = len(g.Image) * 100 // Default 100ms per frame
	}

	// Process each frame
	for frameIdx, frame := range g.Image {
		// Convert frame to RGBA
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, frame, bounds.Min, draw.Src)

		// Calculate visibility times
		startTime := 0
		for i := 0; i < frameIdx; i++ {
			if i < len(g.Delay) && g.Delay[i] > 0 {
				startTime += g.Delay[i] * 10
			} else {
				startTime += 100
			}
		}

		// Create a group for this frame with visibility animation
		fmt.Fprintf(w, "\n<g visibility=\"hidden\">")
		fmt.Fprintf(w, "\n  <animate attributeName=\"visibility\" values=\"hidden;visible;hidden\" ")
		fmt.Fprintf(w, "dur=\"%dms\" begin=\"%dms\" repeatCount=\"indefinite\" />", totalDuration, startTime)

		// Generate shapes for this frame
		if config.pixelTrace {
			pixelPaths := tracePixelBoundaries(rgba)
			generateAnimatedPixelPaths(w, pixelPaths, config)
		} else {
			var optimizedRuns []pixelRun
			if config.pack2D {
				optimizedRuns = find2DRectangles(rgba)
			} else {
				horizontalRuns := findPixelRuns(rgba)
				verticalMerged := mergeRunsVertically(horizontalRuns)

				verticalRuns := findPixelRunsVertical(rgba)
				horizontalMerged := mergeRunsHorizontally(verticalRuns)

				if len(verticalMerged) <= len(horizontalMerged) {
					optimizedRuns = verticalMerged
				} else {
					optimizedRuns = horizontalMerged
				}
			}
			generateAnimatedShapes(w, optimizedRuns, config)
		}

		fmt.Fprintln(w, "\n</g>")
	}

	fmt.Fprintln(w, "\n</svg>")
}

func generateAnimatedPixelPaths(w io.Writer, paths []pixelPath, config optimizationConfig) {
	for _, pixelPath := range paths {
		if len(pixelPath.edges) == 0 {
			continue
		}

		// Build SVG path
		pathStr := ""
		if len(pixelPath.edges) > 0 {
			firstEdge := pixelPath.edges[0]
			pathStr = fmt.Sprintf("M%d %d", firstEdge.start.x, firstEdge.start.y)

			prevEnd := firstEdge.start
			for _, edge := range pixelPath.edges {
				dx := edge.end.x - prevEnd.x
				dy := edge.end.y - prevEnd.y

				if dx == 0 && dy != 0 {
					pathStr += fmt.Sprintf("v%d", dy)
				} else if dy == 0 && dx != 0 {
					pathStr += fmt.Sprintf("h%d", dx)
				} else if dx != 0 || dy != 0 {
					pathStr += fmt.Sprintf("l%d %d", dx, dy)
				}

				prevEnd = edge.end
			}

			pathStr += "z"
		}

		fillColor := fmt.Sprintf("#%02x%02x%02x", pixelPath.color.R, pixelPath.color.G, pixelPath.color.B)
		if pixelPath.color.A < 255 {
			opacity := float64(pixelPath.color.A) / 255.0
			fmt.Fprintf(w, "\n  <path d=\"%s\" fill=\"%s\" opacity=\"%.3f\"/>", pathStr, fillColor, opacity)
		} else {
			fmt.Fprintf(w, "\n  <path d=\"%s\" fill=\"%s\"/>", pathStr, fillColor)
		}
	}
}

func generateAnimatedShapes(w io.Writer, runs []pixelRun, config optimizationConfig) {
	for _, run := range runs {
		if run.color.A == 0 {
			continue
		}

		fillColor := fmt.Sprintf("#%02x%02x%02x", run.color.R, run.color.G, run.color.B)

		if config.usePaths {
			if run.color.A < 255 {
				opacity := float64(run.color.A) / 255.0
				fmt.Fprintf(w, "\n  <path d=\"M%d %dh%dv%dh-%dz\" fill=\"%s\" opacity=\"%.3f\"/>",
					run.x, run.y, run.width, run.height, run.width, fillColor, opacity)
			} else {
				fmt.Fprintf(w, "\n  <path d=\"M%d %dh%dv%dh-%dz\" fill=\"%s\"/>",
					run.x, run.y, run.width, run.height, run.width, fillColor)
			}
		} else {
			if run.color.A < 255 {
				opacity := float64(run.color.A) / 255.0
				fmt.Fprintf(w, "\n  <rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\" opacity=\"%.3f\"/>",
					run.x, run.y, run.width, run.height, fillColor, opacity)
			} else {
				fmt.Fprintf(w, "\n  <rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\"/>",
					run.x, run.y, run.width, run.height, fillColor)
			}
		}
	}
}

func convert(ctx context.Context, cmd *CLI) error {
	// Generate output filename if not provided
	if cmd.Output == "" {
		cmd.Output = strings.TrimSuffix(cmd.Input, filepath.Ext(cmd.Input)) + ".svg"
	}

	// Open input file
	file, err := os.Open(cmd.Input)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Decode image based on file extension
	ext := strings.ToLower(filepath.Ext(cmd.Input))
	switch ext {
	case ".png":
		img, err := png.Decode(file)
		if err != nil {
			return fmt.Errorf("failed to decode PNG: %w", err)
		}
		return convertSingleImage(ctx, cmd, img)
	case ".gif":
		// Seek back to start for GIF decoding
		file.Seek(0, 0)
		g, err := gif.DecodeAll(file)
		if err != nil {
			return fmt.Errorf("failed to decode GIF: %w", err)
		}

		// Check if it's animated
		if len(g.Image) > 1 && (cmd.Animated || cmd.FrameFiles) {
			return convertAnimatedGIF(ctx, cmd, g)
		} else {
			// Just convert the first frame
			return convertSingleImage(ctx, cmd, g.Image[0])
		}
	default:
		return fmt.Errorf("unsupported file format: %s (only .png and .gif are supported)", ext)
	}
}

func convertSingleImage(ctx context.Context, cmd *CLI, img image.Image) error {
	// Convert to RGBA for consistent color handling
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	// Define all optimization combinations to try
	configs := generateOptimizationConfigs()

	// Try each configuration and find the best one
	bestConfig, bestSize, results := findBestOptimization(rgba, bounds, cmd, configs)

	if cmd.Verbose {
		// Find baseline (no optimizations) for comparison
		var baselineSize int
		for _, result := range results {
			if result.config.name == "RLE + rects" {
				baselineSize = result.size
				break
			}
		}

		// Show all results
		log.Printf("Tested %d optimization combinations:", len(results))
		log.Printf("Baseline (no optimizations): %d bytes", baselineSize)

		// Show top 10 results with savings
		log.Println("\nTop 10 results:")
		for i, result := range results {
			if i >= 10 {
				break
			}
			savings := float64(baselineSize-result.size) / float64(baselineSize) * 100
			if savings > 0 {
				log.Printf("  %2d. %s: %d bytes (%.1f%% smaller)", i+1, result.config.name, result.size, savings)
			} else if savings < 0 {
				log.Printf("  %2d. %s: %d bytes (%.1f%% larger)", i+1, result.config.name, result.size, -savings)
			} else {
				log.Printf("  %2d. %s: %d bytes (same size)", i+1, result.config.name, result.size)
			}
		}

		// Show worst result for comparison
		if len(results) > 10 {
			worst := results[len(results)-1]
			worstSavings := float64(baselineSize-worst.size) / float64(baselineSize) * 100
			if worstSavings < 0 {
				log.Printf("  ...\n  %2d. %s: %d bytes (%.1f%% larger, worst)", len(results), worst.config.name, worst.size, -worstSavings)
			} else {
				log.Printf("  ...\n  %2d. %s: %d bytes (worst)", len(results), worst.config.name, worst.size)
			}
		}

		// Show details of the best configuration
		log.Printf("\nBest configuration: %s", bestConfig.name)
		log.Printf("  Size: %d bytes", bestSize)
		bestSavings := float64(baselineSize-bestSize) / float64(baselineSize) * 100
		log.Printf("  Reduction: %.1f%% from baseline (%d → %d bytes)", bestSavings, baselineSize, bestSize)

		if bestConfig.pixelTrace {
			log.Printf("  Method: Pixel boundary tracing")
		} else {
			log.Printf("  Method: Rectangle generation")
			log.Printf("  - 2D packing: %v", bestConfig.pack2D)
			log.Printf("  - Use paths: %v", bestConfig.usePaths)
			log.Printf("  - Group colors: %v", bestConfig.groupColors)
		}
		log.Printf("  - CSS classes: %v", bestConfig.cssColors)
	}

	// Generate the final output using the best configuration
	return generateOptimizedOutput(cmd.Output, rgba, bounds, cmd, bestConfig)
}

func generateOptimizationConfigs() []optimizationConfig {
	var configs []optimizationConfig

	// Generate all possible combinations
	// For pixel trace, we only vary cssColors (other flags don't apply)
	for _, pixelTrace := range []bool{false, true} {
		if pixelTrace {
			// Pixel trace mode - only CSS colors matters
			for _, cssColors := range []bool{false, true} {
				name := "Pixel trace"
				if cssColors {
					name += " + CSS"
				}
				configs = append(configs, optimizationConfig{
					name:        name,
					cssColors:   cssColors,
					groupColors: false,
					usePaths:    false,
					pack2D:      false,
					pixelTrace:  true,
				})
			}
		} else {
			// Rectangle mode - try all combinations
			for _, cssColors := range []bool{false, true} {
				for _, groupColors := range []bool{false, true} {
					for _, usePaths := range []bool{false, true} {
						for _, pack2D := range []bool{false, true} {
							// Build name based on features
							var features []string
							if pack2D {
								features = append(features, "2D pack")
							} else {
								features = append(features, "RLE")
							}
							if usePaths {
								features = append(features, "paths")
							} else {
								features = append(features, "rects")
							}
							if groupColors {
								features = append(features, "grouped")
							}
							if cssColors {
								features = append(features, "CSS")
							}

							name := strings.Join(features, " + ")
							if name == "" {
								name = "No optimizations"
							}

							configs = append(configs, optimizationConfig{
								name:        name,
								cssColors:   cssColors,
								groupColors: groupColors,
								usePaths:    usePaths,
								pack2D:      pack2D,
								pixelTrace:  false,
							})
						}
					}
				}
			}
		}
	}

	return configs
}

type optimizationResult struct {
	config optimizationConfig
	size   int
}

func findBestOptimization(img *image.RGBA, bounds image.Rectangle, cmd *CLI, configs []optimizationConfig) (optimizationConfig, int, []optimizationResult) {
	var results []optimizationResult
	var bestConfig optimizationConfig
	bestSize := int(^uint(0) >> 1) // Max int

	for _, config := range configs {
		// Generate SVG to buffer
		var buf strings.Builder
		generateWithConfig(&buf, img, bounds, cmd, config)

		size := buf.Len()
		results = append(results, optimizationResult{config: config, size: size})

		if size < bestSize {
			bestSize = size
			bestConfig = config
		}
	}

	// Sort results by size for verbose output
	sort.Slice(results, func(i, j int) bool {
		return results[i].size < results[j].size
	})

	return bestConfig, bestSize, results
}

func generateWithConfig(w io.Writer, img *image.RGBA, bounds image.Rectangle, cmd *CLI, config optimizationConfig) {
	// Find shapes using the appropriate algorithm
	var optimizedRuns []pixelRun
	var pixelPaths []pixelPath

	if config.pixelTrace {
		// Use pixel boundary tracing
		pixelPaths = tracePixelBoundaries(img)
	} else if config.pack2D {
		// Use 2D packing algorithm
		optimizedRuns = find2DRectangles(img)
	} else {
		// Use original run-length encoding approach
		horizontalRuns := findPixelRuns(img)
		verticalMerged := mergeRunsVertically(horizontalRuns)

		// Try vertical-first approach
		verticalRuns := findPixelRunsVertical(img)
		horizontalMerged := mergeRunsHorizontally(verticalRuns)

		// Choose the approach with fewer rectangles
		if len(verticalMerged) <= len(horizontalMerged) {
			optimizedRuns = verticalMerged
		} else {
			optimizedRuns = horizontalMerged
		}
	}

	// Create a temporary command structure with the config settings
	tempCmd := &CLI{
		Scale: cmd.Scale,
	}

	// Generate SVG
	if config.pixelTrace && len(pixelPaths) > 0 {
		generatePixelTraceSVGWithConfig(w, pixelPaths, bounds, tempCmd, config)
	} else if config.groupColors || config.cssColors || config.usePaths {
		generateOptimizedSVGWithConfig(w, optimizedRuns, bounds, tempCmd, config)
	} else {
		// Use simple generation
		generateSimpleSVG(w, optimizedRuns, bounds, tempCmd)
	}
}

func generateOptimizedOutput(outputPath string, img *image.RGBA, bounds image.Rectangle, cmd *CLI, config optimizationConfig) error {
	// Create output file
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Generate with best config
	generateWithConfig(out, img, bounds, cmd, config)

	// Get file size
	stat, err := out.Stat()
	if err != nil {
		return err
	}

	// Calculate baseline size for comparison (if not verbose)
	if !cmd.Verbose {
		var buf strings.Builder
		baselineConfig := optimizationConfig{
			name:        "baseline",
			cssColors:   false,
			groupColors: false,
			usePaths:    false,
			pack2D:      false,
			pixelTrace:  false,
		}
		generateWithConfig(&buf, img, bounds, cmd, baselineConfig)
		baselineSize := buf.Len()

		savings := float64(baselineSize-int(stat.Size())) / float64(baselineSize) * 100
		if savings > 0 {
			log.Printf("Generated %s (%d bytes, %.1f%% smaller than baseline)", outputPath, stat.Size(), savings)
		} else {
			log.Printf("Generated %s (%d bytes)", outputPath, stat.Size())
		}
	} else {
		log.Printf("Generated %s using '%s' optimization", outputPath, config.name)
	}

	return nil
}

func findPixelRuns(img *image.RGBA) []pixelRun {
	bounds := img.Bounds()
	var runs []pixelRun

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		var currentRun *pixelRun

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := img.RGBAAt(x, y)

			// Skip transparent pixels
			if pixel.A == 0 {
				if currentRun != nil {
					runs = append(runs, *currentRun)
					currentRun = nil
				}
				continue
			}

			if currentRun == nil || !colorsEqual(currentRun.color, pixel) {
				// Start new run
				if currentRun != nil {
					runs = append(runs, *currentRun)
				}
				currentRun = &pixelRun{
					x:      x,
					y:      y,
					width:  1,
					height: 1,
					color:  pixel,
				}
			} else {
				// Extend current run
				currentRun.width++
			}
		}

		// Add final run if exists
		if currentRun != nil {
			runs = append(runs, *currentRun)
		}
	}

	return runs
}

func mergeRunsVertically(runs []pixelRun) []pixelRun {
	if len(runs) == 0 {
		return runs
	}

	// Create a map to track runs by their key (x, width, color)
	type runKey struct {
		x, width int
		color    color.RGBA
	}

	runMap := make(map[runKey][]pixelRun)

	for _, run := range runs {
		key := runKey{x: run.x, width: run.width, color: run.color}
		runMap[key] = append(runMap[key], run)
	}

	var optimized []pixelRun

	for _, group := range runMap {
		// Sort by y coordinate
		sortedRuns := group
		for i := 0; i < len(sortedRuns)-1; i++ {
			for j := i + 1; j < len(sortedRuns); j++ {
				if sortedRuns[i].y > sortedRuns[j].y {
					sortedRuns[i], sortedRuns[j] = sortedRuns[j], sortedRuns[i]
				}
			}
		}

		// Merge consecutive runs
		merged := []pixelRun{}
		current := sortedRuns[0]

		for i := 1; i < len(sortedRuns); i++ {
			next := sortedRuns[i]

			// Check if runs are adjacent vertically
			if current.y+current.height == next.y {
				// Extend the height
				current.height += next.height
			} else {
				// Save current and start a new one
				merged = append(merged, current)
				current = next
			}
		}
		// Don't forget the last one
		merged = append(merged, current)

		optimized = append(optimized, merged...)
	}

	return optimized
}

func colorsEqual(a, b color.RGBA) bool {
	return a.R == b.R && a.G == b.G && a.B == b.B && a.A == b.A
}

func findPixelRunsVertical(img *image.RGBA) []pixelRun {
	bounds := img.Bounds()
	var runs []pixelRun

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		var currentRun *pixelRun

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pixel := img.RGBAAt(x, y)

			// Skip transparent pixels
			if pixel.A == 0 {
				if currentRun != nil {
					runs = append(runs, *currentRun)
					currentRun = nil
				}
				continue
			}

			if currentRun == nil || !colorsEqual(currentRun.color, pixel) {
				// Start new run
				if currentRun != nil {
					runs = append(runs, *currentRun)
				}
				currentRun = &pixelRun{
					x:      x,
					y:      y,
					width:  1,
					height: 1,
					color:  pixel,
				}
			} else {
				// Extend current run
				currentRun.height++
			}
		}

		// Add final run if exists
		if currentRun != nil {
			runs = append(runs, *currentRun)
		}
	}

	return runs
}

func mergeRunsHorizontally(runs []pixelRun) []pixelRun {
	if len(runs) == 0 {
		return runs
	}

	// Create a map to track runs by their key (y, height, color)
	type runKey struct {
		y, height int
		color     color.RGBA
	}

	runMap := make(map[runKey][]pixelRun)

	for _, run := range runs {
		key := runKey{y: run.y, height: run.height, color: run.color}
		runMap[key] = append(runMap[key], run)
	}

	var optimized []pixelRun

	for _, group := range runMap {
		// Sort by x coordinate
		sortedRuns := group
		for i := 0; i < len(sortedRuns)-1; i++ {
			for j := i + 1; j < len(sortedRuns); j++ {
				if sortedRuns[i].x > sortedRuns[j].x {
					sortedRuns[i], sortedRuns[j] = sortedRuns[j], sortedRuns[i]
				}
			}
		}

		// Merge consecutive runs
		merged := []pixelRun{}
		current := sortedRuns[0]

		for i := 1; i < len(sortedRuns); i++ {
			next := sortedRuns[i]

			// Check if runs are adjacent horizontally
			if current.x+current.width == next.x {
				// Extend the width
				current.width += next.width
			} else {
				// Save current and start a new one
				merged = append(merged, current)
				current = next
			}
		}
		// Don't forget the last one
		merged = append(merged, current)

		optimized = append(optimized, merged...)
	}

	return optimized
}

func generateSimpleSVG(w io.Writer, runs []pixelRun, bounds image.Rectangle, cmd *CLI) {
	baseWidth := bounds.Dx()
	baseHeight := bounds.Dy()
	scaledWidth := baseWidth * cmd.Scale
	scaledHeight := baseHeight * cmd.Scale

	fmt.Fprintf(w, `<?xml version="1.0"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`,
		scaledWidth, scaledHeight, baseWidth, baseHeight)

	for _, run := range runs {
		if run.color.A == 0 {
			continue
		}

		fillColor := fmt.Sprintf("#%02x%02x%02x", run.color.R, run.color.G, run.color.B)

		if run.color.A < 255 {
			opacity := float64(run.color.A) / 255.0
			fmt.Fprintf(w, "\n<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\" opacity=\"%.3f\"/>",
				run.x, run.y, run.width, run.height, fillColor, opacity)
		} else {
			fmt.Fprintf(w, "\n<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\"/>",
				run.x, run.y, run.width, run.height, fillColor)
		}
	}

	fmt.Fprintln(w, "\n</svg>")
}

func generateOptimizedSVGWithConfig(w io.Writer, runs []pixelRun, bounds image.Rectangle, cmd *CLI, config optimizationConfig) {
	baseWidth := bounds.Dx()
	baseHeight := bounds.Dy()
	scaledWidth := baseWidth * cmd.Scale
	scaledHeight := baseHeight * cmd.Scale

	// Write SVG header
	fmt.Fprintf(w, `<?xml version="1.0"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`,
		scaledWidth, scaledHeight, baseWidth, baseHeight)

	// Filter out transparent pixels
	opaqueRuns := []pixelRun{}
	for _, run := range runs {
		if run.color.A > 0 {
			opaqueRuns = append(opaqueRuns, run)
		}
	}

	// Generate CSS classes if needed
	var classMap map[color.RGBA]string
	if config.cssColors {
		classMap = generateCSSColors(w, opaqueRuns)
	}

	if config.groupColors {
		// Group rectangles by color
		generateGroupedRectsWithConfig(w, opaqueRuns, cmd, classMap, config)
	} else {
		// Generate individual elements
		for _, run := range opaqueRuns {
			if config.usePaths {
				generatePathWithConfig(w, run, cmd, classMap, config)
			} else {
				generateRectWithConfig(w, run, cmd, classMap, config)
			}
		}
	}

	fmt.Fprintln(w, "\n</svg>")
}

func generatePixelTraceSVGWithConfig(w io.Writer, paths []pixelPath, bounds image.Rectangle, cmd *CLI, config optimizationConfig) {
	baseWidth := bounds.Dx()
	baseHeight := bounds.Dy()
	scaledWidth := baseWidth * cmd.Scale
	scaledHeight := baseHeight * cmd.Scale

	// Write SVG header
	fmt.Fprintf(w, `<?xml version="1.0"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`,
		scaledWidth, scaledHeight, baseWidth, baseHeight)

	// Generate CSS classes if needed
	var classMap map[color.RGBA]string
	if config.cssColors {
		// Count colors
		colorCount := make(map[color.RGBA]int)
		for _, p := range paths {
			colorCount[p.color]++
		}

		// Sort colors by frequency
		type colorFreq struct {
			color color.RGBA
			count int
		}

		colors := []colorFreq{}
		for c, count := range colorCount {
			colors = append(colors, colorFreq{c, count})
		}

		sort.Slice(colors, func(i, j int) bool {
			return colors[i].count > colors[j].count
		})

		// Generate CSS classes
		fmt.Fprintln(w, "\n<defs><style>")
		classMap = make(map[color.RGBA]string)

		for i, cf := range colors {
			if i >= 26 { // Limit to single letters
				break
			}
			className := string(rune('a' + i))
			classMap[cf.color] = className

			if cf.color.A < 255 {
				fmt.Fprintf(w, ".%s{fill:#%02x%02x%02x;opacity:%.3f}\n",
					className, cf.color.R, cf.color.G, cf.color.B,
					float64(cf.color.A)/255.0)
			} else {
				fmt.Fprintf(w, ".%s{fill:#%02x%02x%02x}\n",
					className, cf.color.R, cf.color.G, cf.color.B)
			}
		}

		fmt.Fprintln(w, "</style></defs>")
	}

	// Generate paths
	for _, pixelPath := range paths {
		if len(pixelPath.edges) == 0 {
			continue
		}

		// Build SVG path with optimized commands
		pathStr := ""
		if len(pixelPath.edges) > 0 {
			// Move to the start of the first edge
			firstEdge := pixelPath.edges[0]
			pathStr = fmt.Sprintf("M%d %d", firstEdge.start.x, firstEdge.start.y)

			// Add optimized path commands
			prevEnd := firstEdge.start
			for _, edge := range pixelPath.edges {
				dx := edge.end.x - prevEnd.x
				dy := edge.end.y - prevEnd.y

				if dx == 0 && dy != 0 {
					pathStr += fmt.Sprintf("v%d", dy)
				} else if dy == 0 && dx != 0 {
					pathStr += fmt.Sprintf("h%d", dx)
				} else if dx != 0 || dy != 0 {
					pathStr += fmt.Sprintf("l%d %d", dx, dy)
				}

				prevEnd = edge.end
			}

			pathStr += "z" // Close the path
		}

		// Write path element
		if config.cssColors && classMap != nil {
			if className, ok := classMap[pixelPath.color]; ok {
				fmt.Fprintf(w, "\n<path d=\"%s\" class=\"%s\"/>", pathStr, className)
			} else {
				fillColor := fmt.Sprintf("#%02x%02x%02x", pixelPath.color.R, pixelPath.color.G, pixelPath.color.B)
				if pixelPath.color.A < 255 {
					opacity := float64(pixelPath.color.A) / 255.0
					fmt.Fprintf(w, "\n<path d=\"%s\" fill=\"%s\" opacity=\"%.3f\"/>", pathStr, fillColor, opacity)
				} else {
					fmt.Fprintf(w, "\n<path d=\"%s\" fill=\"%s\"/>", pathStr, fillColor)
				}
			}
		} else {
			fillColor := fmt.Sprintf("#%02x%02x%02x", pixelPath.color.R, pixelPath.color.G, pixelPath.color.B)
			if pixelPath.color.A < 255 {
				opacity := float64(pixelPath.color.A) / 255.0
				fmt.Fprintf(w, "\n<path d=\"%s\" fill=\"%s\" opacity=\"%.3f\"/>", pathStr, fillColor, opacity)
			} else {
				fmt.Fprintf(w, "\n<path d=\"%s\" fill=\"%s\"/>", pathStr, fillColor)
			}
		}
	}

	fmt.Fprintln(w, "\n</svg>")
}

func generateCSSColors(w io.Writer, runs []pixelRun) map[color.RGBA]string {
	// Count color frequency
	colorCount := make(map[color.RGBA]int)
	for _, run := range runs {
		colorCount[run.color]++
	}

	// Sort colors by frequency
	type colorFreq struct {
		color color.RGBA
		count int
	}

	colors := []colorFreq{}
	for c, count := range colorCount {
		colors = append(colors, colorFreq{c, count})
	}

	sort.Slice(colors, func(i, j int) bool {
		return colors[i].count > colors[j].count
	})

	// Generate CSS classes for top colors
	fmt.Fprintln(w, "\n<defs><style>")
	classMap := make(map[color.RGBA]string)

	for i, cf := range colors {
		if i >= 26 { // Limit to single letters a-z
			break
		}
		className := string(rune('a' + i))
		classMap[cf.color] = className

		if cf.color.A < 255 {
			fmt.Fprintf(w, ".%s{fill:#%02x%02x%02x;opacity:%.3f}\n",
				className, cf.color.R, cf.color.G, cf.color.B,
				float64(cf.color.A)/255.0)
		} else {
			fmt.Fprintf(w, ".%s{fill:#%02x%02x%02x}\n",
				className, cf.color.R, cf.color.G, cf.color.B)
		}
	}

	fmt.Fprintln(w, "</style></defs>")

	return classMap
}

func generateGroupedRectsWithConfig(w io.Writer, runs []pixelRun, cmd *CLI, classMap map[color.RGBA]string, config optimizationConfig) {
	// Group runs by color
	colorGroups := make(map[color.RGBA][]pixelRun)
	for _, run := range runs {
		colorGroups[run.color] = append(colorGroups[run.color], run)
	}

	// Generate groups
	for c, group := range colorGroups {
		// Check if we have a CSS class for this color
		if config.cssColors && classMap != nil {
			if className, ok := classMap[c]; ok {
				fmt.Fprintf(w, "\n<g class=\"%s\">", className)
			} else {
				// Fall back to inline style
				fillColor := fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
				if c.A < 255 {
					opacity := float64(c.A) / 255.0
					fmt.Fprintf(w, "\n<g fill=\"%s\" opacity=\"%.3f\">", fillColor, opacity)
				} else {
					fmt.Fprintf(w, "\n<g fill=\"%s\">", fillColor)
				}
			}
		} else {
			fillColor := fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
			if c.A < 255 {
				opacity := float64(c.A) / 255.0
				fmt.Fprintf(w, "\n<g fill=\"%s\" opacity=\"%.3f\">", fillColor, opacity)
			} else {
				fmt.Fprintf(w, "\n<g fill=\"%s\">", fillColor)
			}
		}

		for _, run := range group {
			if config.usePaths {
				fmt.Fprintf(w, "\n  <path d=\"M%d %dh%dv%dh-%dz\"/>", run.x, run.y, run.width, run.height, run.width)
			} else {
				fmt.Fprintf(w, "\n  <rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\"/>", run.x, run.y, run.width, run.height)
			}
		}

		fmt.Fprintln(w, "\n</g>")
	}
}

func generatePathWithConfig(w io.Writer, run pixelRun, cmd *CLI, classMap map[color.RGBA]string, config optimizationConfig) {
	// Check for CSS class
	if config.cssColors && classMap != nil {
		if className, ok := classMap[run.color]; ok {
			fmt.Fprintf(w, "\n<path d=\"M%d %dh%dv%dh-%dz\" class=\"%s\"/>",
				run.x, run.y, run.width, run.height, run.width, className)
			return
		}
	}

	// Fall back to inline style
	fillColor := fmt.Sprintf("#%02x%02x%02x", run.color.R, run.color.G, run.color.B)

	if run.color.A < 255 {
		opacity := float64(run.color.A) / 255.0
		fmt.Fprintf(w, "\n<path d=\"M%d %dh%dv%dh-%dz\" fill=\"%s\" opacity=\"%.3f\"/>",
			run.x, run.y, run.width, run.height, run.width, fillColor, opacity)
	} else {
		fmt.Fprintf(w, "\n<path d=\"M%d %dh%dv%dh-%dz\" fill=\"%s\"/>",
			run.x, run.y, run.width, run.height, run.width, fillColor)
	}
}

func generateRectWithConfig(w io.Writer, run pixelRun, cmd *CLI, classMap map[color.RGBA]string, config optimizationConfig) {
	// Check for CSS class
	if config.cssColors && classMap != nil {
		if className, ok := classMap[run.color]; ok {
			fmt.Fprintf(w, "\n<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" class=\"%s\"/>",
				run.x, run.y, run.width, run.height, className)
			return
		}
	}

	// Fall back to inline style
	fillColor := fmt.Sprintf("#%02x%02x%02x", run.color.R, run.color.G, run.color.B)

	if run.color.A < 255 {
		opacity := float64(run.color.A) / 255.0
		fmt.Fprintf(w, "\n<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\" opacity=\"%.3f\"/>",
			run.x, run.y, run.width, run.height, fillColor, opacity)
	} else {
		fmt.Fprintf(w, "\n<rect x=\"%d\" y=\"%d\" width=\"%d\" height=\"%d\" fill=\"%s\"/>",
			run.x, run.y, run.width, run.height, fillColor)
	}
}

func find2DRectangles(img *image.RGBA) []pixelRun {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a visited map
	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}

	var rectangles []pixelRun

	// Scan the image for unvisited pixels
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if visited[y-bounds.Min.Y][x-bounds.Min.X] {
				continue
			}

			pixel := img.RGBAAt(x, y)
			if pixel.A == 0 {
				visited[y-bounds.Min.Y][x-bounds.Min.X] = true
				continue
			}

			// Find the largest rectangle starting from this pixel
			rect := findMaxRectangle(img, x, y, pixel, visited, bounds)
			if rect.width > 0 && rect.height > 0 {
				rectangles = append(rectangles, rect)
			}
		}
	}

	return rectangles
}

func findMaxRectangle(img *image.RGBA, startX, startY int, targetColor color.RGBA, visited [][]bool, bounds image.Rectangle) pixelRun {
	// Find maximum width
	maxWidth := 0
	for x := startX; x < bounds.Max.X; x++ {
		if visited[startY-bounds.Min.Y][x-bounds.Min.X] || !colorsEqual(img.RGBAAt(x, startY), targetColor) {
			break
		}
		maxWidth++
	}

	if maxWidth == 0 {
		return pixelRun{}
	}

	// Find maximum height that maintains the width
	maxHeight := 1
	for y := startY + 1; y < bounds.Max.Y; y++ {
		// Check if we can extend the rectangle to this row
		canExtend := true
		for x := startX; x < startX+maxWidth; x++ {
			if visited[y-bounds.Min.Y][x-bounds.Min.X] || !colorsEqual(img.RGBAAt(x, y), targetColor) {
				canExtend = false
				break
			}
		}

		if !canExtend {
			break
		}
		maxHeight++
	}

	// Mark all pixels in the rectangle as visited
	for y := startY; y < startY+maxHeight; y++ {
		for x := startX; x < startX+maxWidth; x++ {
			visited[y-bounds.Min.Y][x-bounds.Min.X] = true
		}
	}

	return pixelRun{
		x:      startX,
		y:      startY,
		width:  maxWidth,
		height: maxHeight,
		color:  targetColor,
	}
}

// Pixel boundary tracing structures and functions
type point struct {
	x, y int
}

type edge struct {
	start, end point
}

type pixelPath struct {
	color color.RGBA
	edges []edge
}

func tracePixelBoundaries(img *image.RGBA) []pixelPath {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Track visited pixels
	visited := make([][]bool, height)
	for i := range visited {
		visited[i] = make([]bool, width)
	}

	var paths []pixelPath

	// Process each unvisited pixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if visited[y-bounds.Min.Y][x-bounds.Min.X] {
				continue
			}

			pixel := img.RGBAAt(x, y)
			if pixel.A == 0 {
				visited[y-bounds.Min.Y][x-bounds.Min.X] = true
				continue
			}

			// Find all connected pixels of this color
			region := floodFill(img, x, y, pixel, visited, bounds)
			if len(region) == 0 {
				continue
			}

			// Trace the boundary of this region
			edges := traceBoundaryEdges(region, pixel, img, bounds)
			if len(edges) > 0 {
				paths = append(paths, pixelPath{
					color: pixel,
					edges: edges,
				})
			}
		}
	}

	return paths
}

func floodFill(img *image.RGBA, startX, startY int, targetColor color.RGBA, visited [][]bool, bounds image.Rectangle) map[point]bool {
	region := make(map[point]bool)
	queue := []point{{startX, startY}}

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]

		if p.x < bounds.Min.X || p.x >= bounds.Max.X || p.y < bounds.Min.Y || p.y >= bounds.Max.Y {
			continue
		}

		if visited[p.y-bounds.Min.Y][p.x-bounds.Min.X] {
			continue
		}

		if !colorsEqual(img.RGBAAt(p.x, p.y), targetColor) {
			continue
		}

		visited[p.y-bounds.Min.Y][p.x-bounds.Min.X] = true
		region[p] = true

		// Add 4-connected neighbors
		queue = append(queue, point{p.x + 1, p.y})
		queue = append(queue, point{p.x - 1, p.y})
		queue = append(queue, point{p.x, p.y + 1})
		queue = append(queue, point{p.x, p.y - 1})
	}

	return region
}

func traceBoundaryEdges(region map[point]bool, targetColor color.RGBA, img *image.RGBA, bounds image.Rectangle) []edge {
	// For each pixel in the region, check its 4 edges
	// An edge is on the boundary if the pixel on the other side is different
	edgeSet := make(map[edge]bool)

	for p := range region {
		x, y := p.x, p.y

		// Check top edge
		if y == bounds.Min.Y || !region[point{x, y - 1}] {
			edgeSet[edge{point{x, y}, point{x + 1, y}}] = true
		}

		// Check right edge
		if x == bounds.Max.X-1 || !region[point{x + 1, y}] {
			edgeSet[edge{point{x + 1, y}, point{x + 1, y + 1}}] = true
		}

		// Check bottom edge
		if y == bounds.Max.Y-1 || !region[point{x, y + 1}] {
			edgeSet[edge{point{x + 1, y + 1}, point{x, y + 1}}] = true
		}

		// Check left edge
		if x == bounds.Min.X || !region[point{x - 1, y}] {
			edgeSet[edge{point{x, y + 1}, point{x, y}}] = true
		}
	}

	// Convert edge set to ordered path
	if len(edgeSet) == 0 {
		return nil
	}

	// Build adjacency map
	adjacency := make(map[point][]point)
	for e := range edgeSet {
		adjacency[e.start] = append(adjacency[e.start], e.end)
		adjacency[e.end] = append(adjacency[e.end], e.start)
	}

	// Find a starting point
	var start point
	for p := range adjacency {
		start = p
		break
	}

	// Trace the path
	var path []edge
	current := start
	var prev point
	visited := make(map[edge]bool)

	for {
		// Find next point
		found := false
		for _, next := range adjacency[current] {
			if next == prev {
				continue
			}

			e := edge{current, next}
			reverseE := edge{next, current}

			if visited[e] || visited[reverseE] {
				continue
			}

			path = append(path, e)
			visited[e] = true
			visited[reverseE] = true
			prev = current
			current = next
			found = true
			break
		}

		if !found || current == start {
			break
		}
	}

	return path
}
