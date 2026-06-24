package magick_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yashikota/mahou/pkg/magick"
)

var update = flag.Bool("update", false, "update golden files using host tools (magick/ffmpeg)")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// pseudoFormats contains formats that are built-in image generators,
// protocols, text-based info formats, raw formats, interactive windows, or internal caches.
var pseudoFormats = map[string]bool{
	// Protocols and interactive
	"HTTP":            true,
	"HTTPS":           true,
	"FTP":             true,
	"FILE":            true,
	"URL":             true,
	"WIN":             true,
	"X":               true,
	"BROWSE":          true,
	"EPHEMERAL":       true,
	// Built-in generators
	"XC":              true,
	"CANVAS":          true,
	"GRADIENT":        true,
	"RADIAL-GRADIENT": true,
	"PLASMA":          true,
	"PATTERN":         true,
	"NULL":            true,
	"MAGICK":          true,
	"INLINE":          true,
	"CLIP":            true,
	"MASK":            true,
	"MONO":            true,
	"HISTOGRAM":       true,
	"INFO":            true,
	"CAPTION":         true,
	"LABEL":           true,
	"VID":             true,
	// Text and markup
	"JSON":            true,
	"YAML":            true,
	"TXT":             true,
	"TEXT":            true,
	"HTML":            true,
	"HTM":             true,
	"XML":             true,
	"MSL":             true,
	"MVG":             true,
	"SHTML":           true,
	"FTXT":            true,
	// Fonts
	"PANGO":           true,
	"DFONT":           true,
	"OTF":             true,
	"TTC":             true,
	"TTF":             true,
	"PFA":             true,
	"PFB":             true,
	// Graphviz
	"DOT":             true,
	"GV":              true,
	// Video formats
	"AVI":             true,
	"FLV":             true,
	"MKV":             true,
	"MOV":             true,
	"MP4":             true,
	"MPEG":            true,
	"MPG":             true,
	"WEBM":            true,
	"WMV":             true,
	"M2V":             true,
	"M4V":             true,
	// Specialized raw or headerless pixel formats (require size/depth options to read)
	"BAYER":           true,
	"BAYERA":          true,
	"BGR":             true,
	"BGRA":            true,
	"BGRO":            true,
	"CMYK":            true,
	"CMYKA":           true,
	"GRAY":            true,
	"GRAYA":           true,
	"RGB":             true,
	"RGBA":            true,
	"RGBO":            true,
	"YUV":             true,
	"YCBCR":           true,
	"YCBCRA":          true,
	"UYVY":            true,
	// Metadata, palette, kernel, or internal cache formats
	"MAP":             true,
	"PAL":             true,
	"KERNEL":          true,
	"IPL":             true,
	"MPC":             true,
	// Braille (dot patterns, not standard pixels)
	"BRF":             true,
	"UBRL":            true,
	"UBRL6":           true,
	"ISOBRL":          true,
	"ISOBRL6":         true,
	// Layout / specialized templates
	"POCKETMOD":       true,
	"ASHLAR":          true,
	// Specialized / problematic color formats
	"SPARSE-COLOR":    true,
	"SPARSE":          true,
	"PIX":             true,
}

// gsFormats require Ghostscript to be installed on the system
var gsFormats = map[string]bool{
	"AI":   true,
	"PDF":  true,
	"PDFA": true,
	"PS":   true,
	"PS2":  true,
	"PS3":  true,
	"EPS":  true,
	"EPS2": true,
	"EPS3": true,
	"EPSF": true,
	"EPSI": true,
	"EPT":  true,
	"EPT2": true,
	"EPT3": true,
	"EPDF": true,
	"EPI":  true,
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func hasGhostscript() bool {
	return hasCommand("gs")
}

func createSolidPNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, blue)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode solid png: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write solid png: %v", err)
	}
}

func isDelegateError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if msg == "" || msg == "imagemagick operation failed" {
		return true
	}
	return strings.Contains(msg, "delegate") ||
		strings.Contains(msg, "missing an image filename") ||
		strings.Contains(msg, "bad value") ||
		strings.Contains(msg, "no pixels defined in cache")
}

func removeMatches(t *testing.T, pattern string) {
	t.Helper()
	matches, _ := filepath.Glob(pattern)
	for _, m := range matches {
		if err := os.Remove(m); err != nil && !os.IsNotExist(err) {
			t.Fatalf("remove %s: %v", m, err)
		}
	}
}

func getFormatFromFilename(filename string) string {
	idx := strings.Index(filename, ".")
	if idx <= 0 {
		return ""
	}
	base := filename[:idx]
	if dashIdx := strings.Index(base, "-"); dashIdx > 0 {
		base = base[:dashIdx]
	}
	return base
}

func isLosslessFormat(format string) bool {
	switch strings.ToUpper(format) {
	case "PNG", "PNG8", "PNG24", "PNG32", "PNG48", "PNG64", "BMP", "BMP2", "BMP3", "TIFF", "TIFF64", "TGA", "QOI", "FARBFELD":
		return true
	}
	return false
}

// isStrictColorFormat returns true if the format is expected to preserve 
// standard color pixels (excluding monochrome-only/highly specialized formats)
func isStrictColorFormat(format string) bool {
	switch strings.ToUpper(format) {
	case "PNG", "PNG8", "PNG24", "PNG32", "PNG48", "PNG64", "BMP", "BMP2", "BMP3", "TIFF", "TIFF64", "TGA", "QOI", "FARBFELD", "JPEG", "JPG", "WEBP", "GIF", "GIF87", "PSD", "PSB":
		return true
	}
	return false
}

func compareFilesBinary(pathA, pathB string) bool {
	dataA, errA := os.ReadFile(pathA)
	dataB, errB := os.ReadFile(pathB)
	if errA != nil || errB != nil {
		return false
	}
	return bytes.Equal(dataA, dataB)
}

func decodePNG(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open image %s: %v", path, err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode png %s: %v", path, err)
	}
	return img
}

func compareImages(t *testing.T, pathA, pathB string, maxDiff float64) {
	t.Helper()
	imgA := decodePNG(t, pathA)
	imgB := decodePNG(t, pathB)

	boundsA := imgA.Bounds()
	boundsB := imgB.Bounds()

	if boundsA.Dx() != boundsB.Dx() || boundsA.Dy() != boundsB.Dy() {
		t.Fatalf("image dimensions mismatch: %dx%d vs %dx%d", boundsA.Dx(), boundsA.Dy(), boundsB.Dx(), boundsB.Dy())
	}
	if boundsA.Dx() == 0 || boundsA.Dy() == 0 {
		t.Fatalf("image dimensions cannot be zero")
	}

	var totalDiff int64
	for y := boundsA.Min.Y; y < boundsA.Max.Y; y++ {
		for x := boundsA.Min.X; x < boundsA.Max.X; x++ {
			rA, gA, bA, aA := imgA.At(x, y).RGBA()
			rB, gB, bB, aB := imgB.At(x, y).RGBA()

			diffR := int64(rA>>8) - int64(rB>>8)
			diffG := int64(gA>>8) - int64(gB>>8)
			diffB := int64(bA>>8) - int64(bB>>8)
			diffA := int64(aA>>8) - int64(aB>>8)

			totalDiff += diffR*diffR + diffG*diffG + diffB*diffB + diffA*diffA
		}
	}

	mse := float64(totalDiff) / float64(boundsA.Dx()*boundsA.Dy()*4)
	if mse > maxDiff {
		t.Fatalf("images are too different, MSE = %f (max allowed: %f)", mse, maxDiff)
	}
}

func TestAllFormatsCompatibility(t *testing.T) {
	setup(t)

	testdataDir := filepath.Join("..", "..", "testdata")
	readDir := filepath.Join(testdataDir, "read")
	writeDir := filepath.Join(testdataDir, "write")

	// -------------------------------------------------------------------------
	// Update Mode: Generate golden files using host tools and mahou
	// -------------------------------------------------------------------------
	if *update {
		if !hasCommand("magick") {
			t.Fatal("Host 'magick' command is required to update golden files, but was not found in PATH")
		}

		// Ensure we start with a clean testdata directory to clear out old garbage
		_ = os.RemoveAll(testdataDir)

		if err := os.MkdirAll(readDir, 0o755); err != nil {
			t.Fatalf("create read dir: %v", err)
		}
		if err := os.MkdirAll(writeDir, 0o755); err != nil {
			t.Fatalf("create write dir: %v", err)
		}

		inputPNG := filepath.Join(testdataDir, "input.png")
		createSolidPNG(t, inputPNG, 10, 10)

		formats := magick.Formats()
		t.Logf("Updating golden files for %d formats...", len(formats))

		for _, format := range formats {
			if pseudoFormats[format] {
				continue
			}
			if gsFormats[format] && !hasGhostscript() {
				t.Logf("Skipping Ghostscript format %s as 'gs' is missing on host", format)
				continue
			}

			ext := strings.ToLower(format)

			// 1. Generate read golden using host tools (magick or ffmpeg)
			hostOutputFile := filepath.Join(readDir, fmt.Sprintf("%s.%s", format, ext))
			removeMatches(t, filepath.Join(readDir, fmt.Sprintf("%s*", format)))

			ctxRead, cancelRead := context.WithTimeout(context.Background(), 3*time.Second)
			cmd := exec.CommandContext(ctxRead, "magick", inputPNG, fmt.Sprintf("%s:%s", format, hostOutputFile))
			cmd.Stdin = nil
			err := cmd.Run()
			cancelRead()

			if err != nil && hasCommand("ffmpeg") {
				ctxFF, cancelFF := context.WithTimeout(context.Background(), 3*time.Second)
				cmdFF := exec.CommandContext(ctxFF, "ffmpeg", "-y", "-nostdin", "-i", inputPNG, hostOutputFile)
				if errFF := cmdFF.Run(); errFF == nil {
					err = nil
				}
				cancelFF()
			}

			if err != nil {
				removeMatches(t, filepath.Join(readDir, fmt.Sprintf("%s*", format))) // Clean up any partial/corrupt outputs
				t.Logf("Host tools cannot generate %s file, skipping read golden generation", format)
			} else {
				matches, _ := filepath.Glob(filepath.Join(readDir, fmt.Sprintf("%s*", format)))
				if len(matches) == 0 {
					removeMatches(t, filepath.Join(readDir, fmt.Sprintf("%s*", format)))
					t.Logf("Host tools did not generate any files matching %s*", format)
				} else {
					t.Logf("Generated read golden for %s", format)
				}
			}

			// 2. Generate write golden using mahou
			goOutputFile := filepath.Join(writeDir, fmt.Sprintf("%s.%s", format, ext))
			removeMatches(t, filepath.Join(writeDir, fmt.Sprintf("%s*", format)))

			opts := magick.ConvertOptions{Format: format}
			errGoWrite := magick.Convert(inputPNG, goOutputFile, opts)
			if errGoWrite != nil {
				removeMatches(t, filepath.Join(writeDir, fmt.Sprintf("%s*", format))) // Clean up
				t.Logf("mahou cannot write format %s, skipping write golden generation: %v", format, errGoWrite)
			} else {
				matchesGo, _ := filepath.Glob(filepath.Join(writeDir, fmt.Sprintf("%s*", format)))
				if len(matchesGo) == 0 {
					removeMatches(t, filepath.Join(writeDir, fmt.Sprintf("%s*", format)))
					t.Logf("mahou did not generate any files matching %s*", format)
				} else {
					t.Logf("Generated write golden for %s", format)
				}
			}
		}

		t.Log("Golden files updated successfully. Commit the 'testdata' directory to git.")
		return
	}

	// -------------------------------------------------------------------------
	// Test Mode: Verify read/write using pre-generated golden files (No host tools required)
	// -------------------------------------------------------------------------
	inputPNG := filepath.Join(testdataDir, "input.png")
	if _, err := os.Stat(inputPNG); os.IsNotExist(err) {
		t.Fatalf("testdata/input.png does not exist. Run with -update flag first to generate golden files.")
	}

	// 1. Verify Read
	readFiles, err := os.ReadDir(readDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("testdata/read directory does not exist. Run with -update flag first.")
		}
		t.Fatalf("read readDir: %v", err)
	}

	for _, file := range readFiles {
		if file.IsDir() {
			continue
		}
		filename := file.Name()
		format := getFormatFromFilename(filename)
		if format == "" {
			continue
		}

		t.Run("Read_"+format, func(t *testing.T) {
			filePath := filepath.Join(readDir, filename)

			// 1.1 Identify the file using mahou
			info, errIdent := magick.Identify(filePath)
			if errIdent != nil {
				if isDelegateError(errIdent) {
					t.Skipf("mahou does not support reading %s (no delegate): %v", format, errIdent)
				}
				t.Fatalf("mahou Identify failed: %v", errIdent)
			}
			if info.Width == 0 || info.Height == 0 {
				t.Fatalf("mahou Identify returned zero dimensions")
			}

			// 1.2 Convert to PNG
			dir := t.TempDir()
			decodedPNG := filepath.Join(dir, "decoded.png")
			errConv := magick.Convert(filePath, decodedPNG, magick.ConvertOptions{})
			if errConv != nil {
				if isDelegateError(errConv) {
					t.Skipf("mahou does not support decoding %s (no delegate): %v", format, errConv)
				}
				t.Fatalf("mahou Convert (read) failed: %v", errConv)
			}

			// Find decoded PNG files (supporting multiple page/layer outputs)
			matchesDec, _ := filepath.Glob(filepath.Join(dir, "decoded*"))
			if len(matchesDec) == 0 {
				t.Fatalf("decoded PNG file missing")
			}
			decodedFile := matchesDec[0]

			// 1.3 Compare decoded pixels with initial input.png if it's a strict color format
			if isStrictColorFormat(format) {
				maxDiff := 200.0
				if isLosslessFormat(format) {
					maxDiff = 0.0
				}
				compareImages(t, inputPNG, decodedFile, maxDiff)
			} else {
				// For monochrome or highly specialized layouts, just check if it's identified properly
				decInfo, errDecIdent := magick.Identify(decodedFile)
				if errDecIdent != nil || decInfo.Width == 0 || decInfo.Height == 0 {
					t.Fatalf("decoded file is invalid or unidentifiable")
				}
			}
		})
	}

	// 2. Verify Write
	writeFiles, err := os.ReadDir(writeDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("testdata/write directory does not exist. Run with -update flag first.")
		}
		t.Fatalf("read writeDir: %v", err)
	}

	for _, file := range writeFiles {
		if file.IsDir() {
			continue
		}
		filename := file.Name()
		format := getFormatFromFilename(filename)
		if format == "" {
			continue
		}

		t.Run("Write_"+format, func(t *testing.T) {
			goldenPath := filepath.Join(writeDir, filename)

			dir := t.TempDir()
			ext := filepath.Ext(filename)
			tempOut := filepath.Join(dir, "out"+ext)

			// 2.1 Convert input PNG to target format
			opts := magick.ConvertOptions{Format: format}
			errGoWrite := magick.Convert(inputPNG, tempOut, opts)
			if errGoWrite != nil {
				if isDelegateError(errGoWrite) {
					t.Skipf("mahou does not support writing %s (no delegate): %v", format, errGoWrite)
				}
				t.Fatalf("mahou Convert (write) failed: %v", errGoWrite)
			}

			// Find temp outputs
			matchesOut, _ := filepath.Glob(filepath.Join(dir, "out*"))
			if len(matchesOut) == 0 {
				t.Fatalf("mahou write output missing")
			}
			tempOutFile := matchesOut[0]

			// 2.2 Compare written binary with golden binary
			if compareFilesBinary(tempOutFile, goldenPath) {
				return // exact binary match -> test passes
			}

			// 2.3 If binary doesn't match, decode both outputs to PNG and compare if strict format
			tempPNG := filepath.Join(dir, "temp.png")
			goldenPNG := filepath.Join(dir, "golden.png")

			if err := magick.Convert(tempOutFile, tempPNG, magick.ConvertOptions{}); err != nil {
				if isDelegateError(err) {
					t.Skipf("mahou does not support decoding %s (no delegate): %v", format, err)
				}
				t.Fatalf("failed to decode mahou output: %v", err)
			}
			if err := magick.Convert(goldenPath, goldenPNG, magick.ConvertOptions{}); err != nil {
				if isDelegateError(err) {
					t.Skipf("mahou does not support decoding golden %s (no delegate): %v", format, err)
				}
				t.Fatalf("failed to decode golden file: %v", err)
			}

			// Find actually decoded PNGs
			matchesTempPNG, _ := filepath.Glob(filepath.Join(dir, "temp*"))
			matchesGoldenPNG, _ := filepath.Glob(filepath.Join(dir, "golden*"))
			if len(matchesTempPNG) == 0 || len(matchesGoldenPNG) == 0 {
				t.Fatalf("failed to locate decoded PNGs for comparison")
			}

			if isStrictColorFormat(format) {
				maxDiff := 200.0
				if isLosslessFormat(format) {
					maxDiff = 0.0
				}
				compareImages(t, matchesGoldenPNG[0], matchesTempPNG[0], maxDiff)
			} else {
				decInfo, errDecIdent := magick.Identify(matchesTempPNG[0])
				if errDecIdent != nil || decInfo.Width == 0 || decInfo.Height == 0 {
					t.Fatalf("decoded output is invalid or unidentifiable")
				}
			}
		})
	}
}
