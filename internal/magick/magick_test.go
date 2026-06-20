package magick_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/yashikota/magick-go/internal/magick"
	"github.com/yashikota/magick-go/internal/runtimebundle"
)

var testBundle *runtimebundle.Bundle

func setup(t *testing.T) {
	t.Helper()
	if testBundle == nil {
		bundle, err := runtimebundle.Ensure()
		if err != nil {
			t.Skipf("runtime bundle not available: %v", err)
		}
		policyDir, err := runtimebundle.ApplyPolicy(true)
		if err != nil {
			t.Fatalf("apply policy: %v", err)
		}
		t.Cleanup(func() { os.RemoveAll(policyDir) })
		// ConfigureEnvironment must be called BEFORE Load so that
		// MagickWandGenesis picks up MAGICK_CODER_MODULE_PATH etc.
		runtimebundle.ConfigureEnvironment(bundle.Root, policyDir)
		if _, err := magick.Load(bundle.Root); err != nil {
			t.Fatalf("load magick: %v", err)
		}
		testBundle = bundle
	}
}

func createTestPNG(t *testing.T, path string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test png: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write test png: %v", err)
	}
}

func TestIdentify(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	input := filepath.Join(dir, "test.png")
	createTestPNG(t, input)

	info, err := magick.Identify(input)
	if err != nil {
		t.Fatalf("Identify: %v", err)
	}
	if info.Width != 4 || info.Height != 4 {
		t.Errorf("expected 4x4, got %dx%d", info.Width, info.Height)
	}
	if info.Format != "PNG" {
		t.Errorf("expected format PNG, got %s", info.Format)
	}
}

func TestResize(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	input := filepath.Join(dir, "input.png")
	output := filepath.Join(dir, "resized.png")
	createTestPNG(t, input)

	err := magick.Resize(input, output, 2, magick.ConvertOptions{})
	if err != nil {
		t.Fatalf("Resize: %v", err)
	}

	info, err := magick.Identify(output)
	if err != nil {
		t.Fatalf("Identify resized: %v", err)
	}
	if info.Width != 2 {
		t.Errorf("expected width 2, got %d", info.Width)
	}
}

// TestConvertFormats tests writing to every writable image format supported by
// the bundled ImageMagick runtime. This validates that the delegate libraries
// are correctly linked and functional, not just registered.
func TestConvertFormats(t *testing.T) {
	setup(t)

	type formatSpec struct {
		name      string
		ext       string
		magic     []byte
		linuxOnly bool
	}

	formats := []formatSpec{
		// Raster: standard
		{name: "JPEG", ext: "jpg", magic: []byte{0xFF, 0xD8}},
		{name: "PNG", ext: "png", magic: []byte{0x89, 0x50, 0x4E, 0x47}},
		{name: "PNG8", ext: "png", magic: []byte{0x89, 0x50, 0x4E, 0x47}},
		{name: "PNG24", ext: "png", magic: []byte{0x89, 0x50, 0x4E, 0x47}},
		{name: "PNG32", ext: "png", magic: []byte{0x89, 0x50, 0x4E, 0x47}},
		{name: "PNG48", ext: "png", magic: []byte{0x89, 0x50, 0x4E, 0x47}},
		{name: "PNG64", ext: "png", magic: []byte{0x89, 0x50, 0x4E, 0x47}},
		{name: "WebP", ext: "webp", magic: []byte("RIFF")},
		{name: "TIFF", ext: "tiff"},
		{name: "GIF", ext: "gif", magic: []byte("GIF8")},
		{name: "GIF87", ext: "gif", magic: []byte("GIF8")},
		{name: "BMP", ext: "bmp", magic: []byte("BM")},
		{name: "BMP2", ext: "bmp", magic: []byte("BM")},
		{name: "BMP3", ext: "bmp", magic: []byte("BM")},

		// Modern formats
		{name: "JXL", ext: "jxl", linuxOnly: true},
		{name: "QOI", ext: "qoi", magic: []byte("qoif")},

		// Professional / cinema
		{name: "EXR", ext: "exr", magic: []byte{0x76, 0x2F, 0x31, 0x01}},
		{name: "PSD", ext: "psd", magic: []byte("8BPS")},
		{name: "PSB", ext: "psb", magic: []byte("8BPS")},
		{name: "DPX", ext: "dpx"},
		{name: "CIN", ext: "cin"},
		{name: "FITS", ext: "fits"},
		{name: "HDR", ext: "hdr"},
		{name: "MIFF", ext: "miff"},

		// Legacy / interchange
		{name: "TGA", ext: "tga"},
		{name: "ICO", ext: "ico"},
		{name: "CUR", ext: "cur"},
		{name: "PCX", ext: "pcx"},
		{name: "SGI", ext: "sgi"},
		{name: "SUN", ext: "sun"},
		{name: "XBM", ext: "xbm"},
		{name: "XPM", ext: "xpm"},
		{name: "WBMP", ext: "wbmp"},
		{name: "OTB", ext: "otb"},
		{name: "PALM", ext: "palm"},
		{name: "PICON", ext: "xpm"},
		{name: "RAS", ext: "ras"},
		{name: "VIFF", ext: "viff"},

		// Netpbm family
		{name: "PAM", ext: "pam"},
		{name: "PBM", ext: "pbm"},
		{name: "PGM", ext: "pgm"},
		{name: "PPM", ext: "ppm"},
		{name: "PNM", ext: "pnm"},
		{name: "PFM", ext: "pfm"},
		{name: "PHM", ext: "phm"},

		// Fax
		{name: "FAX", ext: "fax"},
		{name: "G3", ext: "g3"},
		{name: "G4", ext: "g4"},

		// Miscellaneous writable
		{name: "FARBFELD", ext: "ff"},
		{name: "AAI", ext: "aai"},
		{name: "AVS", ext: "avs"},
		{name: "DCX", ext: "dcx"},
		{name: "DDS", ext: "dds"},
		{name: "FL32", ext: "fl32"},
		{name: "FTXT", ext: "ftxt"},
		{name: "HRZ", ext: "hrz"},
		{name: "IPL", ext: "ipl"},
		{name: "MPC", ext: "mpc"},
		{name: "MTV", ext: "mtv"},
		{name: "RGF", ext: "rgf"},
		{name: "SIXEL", ext: "sixel"},
		{name: "VIPS", ext: "vips"},
		{name: "VICAR", ext: "vicar"},
		{name: "MNG", ext: "mng"},
		{name: "JNG", ext: "jng"},

		// JP2/JPEG2000 family
		{name: "JP2", ext: "jp2"},
		{name: "J2K", ext: "j2k"},
		{name: "JPC", ext: "jpc"},

		// TIFF variants
		{name: "PTIF", ext: "ptif"},
		{name: "TIFF64", ext: "tiff"},

		// Braille
		{name: "UBRL", ext: "ubrl"},
		{name: "UBRL6", ext: "ubrl6"},
		{name: "ISOBRL", ext: "isobrl"},
		{name: "ISOBRL6", ext: "isobrl6"},

		// Text/data
		{name: "TXT", ext: "txt"},
		{name: "JSON", ext: "json"},
		{name: "YAML", ext: "yaml"},

		// JBIG (Linux only)
		{name: "JBIG", ext: "jbig", linuxOnly: true},
		{name: "JBG", ext: "jbg", linuxOnly: true},
		{name: "BIE", ext: "bie", linuxOnly: true},
	}

	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			if f.linuxOnly && runtime.GOOS != "linux" {
				t.Skipf("%s not supported on %s", f.name, runtime.GOOS)
			}

			dir := t.TempDir()
			input := filepath.Join(dir, "input.png")
			output := filepath.Join(dir, "output."+f.ext)
			createTestPNG(t, input)

			opts := magick.ConvertOptions{}
			if f.name != f.ext {
				opts.Format = f.name
			}

			err := magick.Convert(input, output, opts)
			if err != nil {
				t.Fatalf("Convert PNG->%s: %v", f.name, err)
			}

			data, err := os.ReadFile(output)
			if err != nil {
				t.Fatalf("read output: %v", err)
			}
			if len(data) == 0 {
				t.Fatalf("output file is empty")
			}
			if f.magic != nil && len(data) >= len(f.magic) {
				if !bytes.HasPrefix(data, f.magic) {
					t.Errorf("magic mismatch: got %x, want %x", data[:len(f.magic)], f.magic)
				}
			}
		})
	}
}

// TestRoundTrip verifies that converting to a format and back preserves image dimensions.
func TestRoundTrip(t *testing.T) {
	setup(t)

	formats := []struct {
		name      string
		ext       string
		linuxOnly bool
	}{
		{"JPEG", "jpg", false},
		{"PNG", "png", false},
		{"WebP", "webp", false},
		{"TIFF", "tiff", false},
		{"GIF", "gif", false},
		{"BMP", "bmp", false},
		{"JXL", "jxl", true},
		{"EXR", "exr", false},
		{"TGA", "tga", false},
		{"PPM", "ppm", false},
		{"PAM", "pam", false},
		{"SGI", "sgi", false},
		{"PCX", "pcx", false},
		{"FARBFELD", "ff", false},
		{"QOI", "qoi", false},
		{"JP2", "jp2", false},
		{"MIFF", "miff", false},
		{"DPX", "dpx", false},
	}

	for _, f := range formats {
		t.Run(f.name, func(t *testing.T) {
			if f.linuxOnly && runtime.GOOS != "linux" {
				t.Skipf("%s not supported on %s", f.name, runtime.GOOS)
			}
			dir := t.TempDir()
			input := filepath.Join(dir, "input.png")
			mid := filepath.Join(dir, "mid."+f.ext)
			output := filepath.Join(dir, "output.png")
			createTestPNG(t, input)

			if err := magick.Convert(input, mid, magick.ConvertOptions{}); err != nil {
				t.Fatalf("Convert PNG->%s: %v", f.name, err)
			}
			if err := magick.Convert(mid, output, magick.ConvertOptions{}); err != nil {
				t.Fatalf("Convert %s->PNG: %v", f.name, err)
			}

			info, err := magick.Identify(output)
			if err != nil {
				t.Fatalf("Identify: %v", err)
			}
			if info.Width != 4 || info.Height != 4 {
				t.Errorf("round-trip lost dimensions: got %dx%d", info.Width, info.Height)
			}
		})
	}
}

func TestFormats(t *testing.T) {
	setup(t)
	formats := magick.Formats()
	if len(formats) == 0 {
		t.Fatal("Formats returned empty list")
	}

	required := []string{
		"JPEG", "PNG", "WEBP", "TIFF", "GIF", "BMP",
		"HEIC", "HEIF", "AVIF",
		"SVG", "SVGZ", "PDF",
		"EXR", "PSD", "PSB",
		"JP2", "J2K", "JPC",
		"DPX", "HDR", "TGA",
		"ICO", "PCX", "SGI", "WBMP",
		"PAM", "PBM", "PGM", "PPM", "PNM",
		"FARBFELD", "QOI",
		"MIFF", "MNG", "JNG",
		"APNG", "PJPEG",
		"FAX", "G3", "G4",
	}
	if runtime.GOOS == "linux" {
		required = append(required, "JXL")
	}
	formatSet := make(map[string]struct{}, len(formats))
	for _, f := range formats {
		formatSet[f] = struct{}{}
	}
	for _, f := range required {
		if _, ok := formatSet[f]; !ok {
			t.Errorf("required format %s not found in supported formats", f)
		}
	}

	t.Logf("total formats: %d", len(formats))
}
