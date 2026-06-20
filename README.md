# magick-go

`magickgo` is a standalone, pure-Go CLI that bundles a complete ImageMagick 7 runtime.
No system dependencies required — the binary extracts a self-contained
`runtime-<target>.tar.zst` into the user cache, configures the ImageMagick
environment, and loads `libMagickWand` dynamically through `purego` (no CGO).

## Supported Targets

| OS    | Architecture | Target          |
|-------|-------------|-----------------|
| Linux | amd64       | `linux-amd64`   |
| Linux | arm64       | `linux-arm64`   |
| macOS | arm64       | `darwin-arm64`  |

## Supported Image Formats

ImageMagick 7.1.2-8 Q16-HDRI with the following delegates enabled:

### Modern Image Formats

| Format | Extension(s)         | Read | Write | Delegate     | Notes                         |
|--------|---------------------|------|-------|-------------|-------------------------------|
| JPEG   | `.jpg`, `.jpeg`     | ✓    | ✓     | libjpeg     | Baseline & progressive        |
| PNG    | `.png`              | ✓    | ✓     | libpng      | 8/16/32/48/64-bit variants    |
| APNG   | `.apng`             | ✓    | ✓     | libpng      | Animated PNG                  |
| WebP   | `.webp`             | ✓    | ✓     | libwebp     | Lossy & lossless              |
| TIFF   | `.tiff`, `.tif`     | ✓    | ✓     | libtiff     | Including BigTIFF (TIFF64)    |
| GIF    | `.gif`              | ✓    | ✓     | built-in    | Animated support              |
| BMP    | `.bmp`              | ✓    | ✓     | built-in    | BMP2/BMP3 variants            |
| HEIC   | `.heic`, `.heif`    | ✓    | ✓     | libheif     | HEVC codec                    |
| AVIF   | `.avif`             | ✓    | ✓     | libheif+aom | AV1 Image Format              |
| JXL    | `.jxl`              | ✓    | ✓     | libjxl      | JPEG XL                       |
| QOI    | `.qoi`              | ✓    | ✓     | built-in    | Quite OK Image                |

### Vector & Document Formats

| Format | Extension(s)         | Read | Write | Delegate     | Notes                         |
|--------|---------------------|------|-------|-------------|-------------------------------|
| SVG    | `.svg`, `.svgz`     | ✓    | ✓     | librsvg     | Rasterized on read            |
| PDF    | `.pdf`, `.pdfa`     | ✓    | ✓     | ghostscript | Multi-page support            |
| EPS    | `.eps`, `.epsf`     | ✓    | ✓     | ghostscript | Encapsulated PostScript       |
| PS     | `.ps`               | ✓    | ✓     | ghostscript | PostScript Level 2/3          |

### Professional & Cinema Formats

| Format | Extension(s)         | Read | Write | Delegate    | Notes                         |
|--------|---------------------|------|-------|------------|-------------------------------|
| EXR    | `.exr`              | ✓    | ✓     | openexr    | HDR, multi-channel            |
| PSD    | `.psd`, `.psb`      | ✓    | ✓     | built-in   | Photoshop (incl. Large PSB)   |
| DPX    | `.dpx`              | ✓    | ✓     | built-in   | SMPTE 268M digital cinema     |
| CIN    | `.cin`              | ✓    | ✓     | built-in   | Kodak Cineon                  |
| HDR    | `.hdr`              | ✓    | ✓     | built-in   | Radiance RGBE                 |
| FITS   | `.fits`, `.fts`     | ✓    | ✓     | built-in   | Astronomy / scientific        |
| MIFF   | `.miff`             | ✓    | ✓     | built-in   | ImageMagick native            |

### JPEG 2000 Family

| Format | Extension(s)         | Read | Write | Delegate    | Notes                         |
|--------|---------------------|------|-------|------------|-------------------------------|
| JP2    | `.jp2`              | ✓    | ✓     | openjp2    | JPEG 2000 Part 1              |
| J2K    | `.j2k`, `.j2c`      | ✓    | ✓     | openjp2    | JPEG 2000 codestream          |
| JPC    | `.jpc`              | ✓    | ✓     | openjp2    | JPEG 2000 codestream          |
| JPM    | `.jpm`              | ✓    | ✓     | openjp2    | JPEG 2000 compound            |

### Legacy & Interchange Formats

| Format   | Extension(s)     | Read | Write | Notes                         |
|----------|-----------------|------|-------|-------------------------------|
| TGA      | `.tga`, `.icb`  | ✓    | ✓     | Targa / Truevision            |
| ICO      | `.ico`          | ✓    | ✓     | Windows icon                  |
| CUR      | `.cur`          | ✓    | ✓     | Windows cursor                |
| PCX      | `.pcx`, `.dcx`  | ✓    | ✓     | PC Paintbrush (multi-page)    |
| SGI      | `.sgi`          | ✓    | ✓     | Silicon Graphics IRIS         |
| SUN      | `.sun`, `.ras`  | ✓    | ✓     | Sun Rasterfile                |
| XBM      | `.xbm`          | ✓    | ✓     | X11 bitmap                    |
| XPM      | `.xpm`          | ✓    | ✓     | X11 pixmap                    |
| WBMP     | `.wbmp`         | ✓    | ✓     | Wireless bitmap               |
| PALM     | `.palm`         | ✓    | ✓     | Palm pixmap                   |
| PICT     | `.pict`, `.pct` | ✓    | ✓     | Apple QuickDraw               |
| VIFF     | `.viff`         | ✓    | ✓     | Khoros Visualization          |
| MNG      | `.mng`          | ✓    | ✓     | Multiple-image PNG            |
| JNG      | `.jng`          | ✓    | ✓     | JPEG Network Graphics         |
| DDS      | `.dds`          | ✓    | ✓     | DirectDraw Surface (DXT1/5)   |
| OTB      | `.otb`          | ✓    | ✓     | On-the-air bitmap             |
| WPG      | `.wpg`          | ✓    | ✓     | WordPerfect Graphics          |

### Netpbm / Portable Pixmap Family

| Format | Extension(s)     | Read | Write | Notes                         |
|--------|-----------------|------|-------|-------------------------------|
| PBM    | `.pbm`          | ✓    | ✓     | Portable bitmap (1-bit)       |
| PGM    | `.pgm`          | ✓    | ✓     | Portable graymap              |
| PPM    | `.ppm`          | ✓    | ✓     | Portable pixmap               |
| PNM    | `.pnm`          | ✓    | ✓     | Portable anymap               |
| PAM    | `.pam`          | ✓    | ✓     | Portable arbitrary map        |
| PFM    | `.pfm`          | ✓    | ✓     | Portable float map            |
| PHM    | `.phm`          | ✓    | ✓     | Portable half-float map       |

### Fax & Braille

| Format   | Extension(s)   | Read | Write | Notes                       |
|----------|---------------|------|-------|-----------------------------|
| FAX      | `.fax`        | ✓    | ✓     | Group 3 fax                 |
| G3       | `.g3`         | ✓    | ✓     | CCITT Group 3               |
| G4       | `.g4`         | ✓    | ✓     | CCITT Group 4               |
| UBRL     | `.ubrl`       | ✓    | ✓     | Unicode braille             |
| ISOBRL   | `.isobrl`     | ✓    | ✓     | ISO/TR 11548-1 braille      |

### Miscellaneous

| Format   | Extension(s)   | Read | Write | Notes                       |
|----------|---------------|------|-------|-----------------------------|
| FARBFELD | `.ff`         | ✓    | ✓     | suckless image format       |
| AAI      | `.aai`        | ✓    | ✓     | Dune HD media player        |
| AVS      | `.avs`        | ✓    | ✓     | AVS X image                 |
| FL32     | `.fl32`       | ✓    | ✓     | 32-bit float pixels         |
| SIXEL    | `.sixel`      | ✓    | ✓     | DEC terminal graphics       |
| VIPS     | `.vips`       | ✓    | ✓     | VIPS image format           |
| MTV      | `.mtv`        | ✓    | ✓     | MTV Raytracer               |
| VICAR    | `.vicar`      | ✓    | ✓     | NASA/JPL VICAR              |
| RGF      | `.rgf`        | ✓    | ✓     | LEGO MINDSTORMS EV3         |
| HRZ      | `.hrz`        | ✓    | ✓     | Slow-scan TV                |
| IPL      | `.ipl`        | ✓    | ✓     | IPLab image                 |
| MPC      | `.mpc`        | ✓    | ✓     | Magick Pixel Cache          |

### Text & Data Output

| Format | Extension(s)   | Read | Write | Notes                        |
|--------|---------------|------|-------|------------------------------|
| TXT    | `.txt`        | ✓    | ✓     | Pixel enumeration            |
| JSON   | `.json`       | —    | ✓     | Image metadata as JSON       |
| YAML   | `.yaml`       | —    | ✓     | Image metadata as YAML       |

### Camera RAW (Read-only)

| Format | Extension(s)                              | Notes                        |
|--------|------------------------------------------|------------------------------|
| DNG    | `.dng`                                   | Adobe Digital Negative       |
| CR2    | `.cr2`, `.cr3`, `.crw`                   | Canon RAW                    |
| NEF    | `.nef`, `.nrw`                           | Nikon RAW                    |
| ARW    | `.arw`                                   | Sony RAW                     |
| ORF    | `.orf`                                   | Olympus RAW                  |
| RAF    | `.raf`                                   | Fujifilm RAW                 |
| RW2    | `.rw2`                                   | Panasonic RAW                |
| PEF    | `.pef`                                   | Pentax RAW                   |
| ERF    | `.erf`                                   | Epson RAW                    |
| SRW    | `.srw`, `.sr2`, `.srf`                   | Samsung RAW                  |
| KDC    | `.kdc`, `.k25`                           | Kodak RAW                    |
| MOS    | `.mos`                                   | Leaf RAW                     |
| MEF    | `.mef`                                   | Mamiya RAW                   |
| IIQ    | `.iiq`                                   | Phase One RAW                |
| 3FR    | `.3fr`                                   | Hasselblad RAW               |
| X3F    | `.x3f`                                   | Sigma RAW                    |
| MDC    | `.mdc`                                   | Minolta RAW                  |
| DCR    | `.dcr`                                   | Kodak RAW                    |

### Platform-specific Formats

| Format | Linux | macOS | Notes                         |
|--------|-------|-------|-------------------------------|
| DJVU   | ✓     | —     | DjVu (requires djvulibre)     |
| JBIG   | ✓     | —     | JBIG1 compression             |
| WMF    | ✓     | —     | Windows Metafile              |
| FFTW   | ✓     | —     | Fourier transform             |
| RAW    | ✓     | —     | libraw camera RAW processing  |

### Delegates

| Platform | Delegates                                                                                              |
|----------|-------------------------------------------------------------------------------------------------------|
| Linux    | bzlib cairo djvu fftw fontconfig freetype heic jbig jng jp2 jpeg jxl lcms lqr lzma openexr pango png ps raqm raw rsvg tiff webp wmf xml zip zlib zstd |
| macOS    | bzlib cairo fontconfig freetype heic jng jp2 jpeg jxl lcms lzma openexr pango png ps rsvg tiff webp xml zlib zstd |

### Format Limitations

| Format | Limitation | Reason |
|--------|-----------|--------|
| HEIC/HEIF/AVIF | Write requires CLI mode | Coder module (`heic.so`) needs `libheif` loaded via dynamic linker at process start; in-process purego binding may not resolve the delegate |
| PDF/EPS/PS | Blocked by default policy | Ghostscript delegate is security-sensitive; use `--policy permissive` to enable |
| SVG (write) | Requires `potrace` | SVG vectorization needs external `potrace` binary (not bundled) |
| DJVU/WMF | Read-only | No encode delegate exists for these formats |
| Camera RAW | Read-only | libraw provides decode only |

### Security Policy

PDF, PostScript (PS/EPS), MVG, and MSL formats are disabled by the default
security policy. URL/HTTP/HTTPS delegates are also blocked. Use
`--policy permissive` to enable all formats for trusted workflows.

### Test Coverage

CI runs actual image I/O tests (not just format registration checks) for **75+ writable formats**:

- **Format conversion test**: Converts a valid PNG to each target format and verifies non-empty output with correct magic bytes
- **Round-trip test**: PNG → format → PNG, verifies dimensions are preserved (18 formats)
- **Format registration test**: Verifies all expected formats are reported by ImageMagick's coder registry

Formats excluded from write tests due to environmental constraints (HEIC/AVIF encode delegate, PDF policy, SVG external tool) are still validated via `doctor --verbose` in CI which confirms the delegate libraries are linked and the format is registered for read.

## Commands

```sh
magickgo doctor --verbose     # Show runtime info, delegates, format support
magickgo formats              # List all supported formats
magickgo identify input.png   # Image metadata (dimensions, format, depth)
magickgo convert input.heic output.webp              # Format conversion
magickgo convert input.png output.jpg --quality 85   # With quality
magickgo convert input.jpg output.jpg --strip        # Strip metadata
magickgo resize input.jpg output.webp --width 1200   # Resize (aspect-ratio preserved)
```

### Options

| Flag             | Description                              |
|------------------|------------------------------------------|
| `--quality N`    | Output quality (1-100, format-dependent) |
| `--strip`        | Remove EXIF/metadata                     |
| `--auto-orient`  | Auto-rotate based on EXIF orientation    |
| `--format FMT`   | Override output format                   |
| `--width N`      | Target width for resize (Lanczos filter) |
| `--policy P`     | `safe` (default) or `permissive`         |
| `--json`         | JSON output (doctor, identify, formats)  |
| `--verbose`      | Verbose output (doctor)                  |

## Architecture

```
┌─────────────────────────────────────┐
│         magickgo binary             │
│  (CGO_ENABLED=0, pure Go)          │
├─────────────────────────────────────┤
│  cmd/magickgo/     CLI commands     │
│  internal/magick/  purego bindings  │
│  internal/runtimebundle/            │
│    ├── embed.go    //go:embed       │
│    ├── extract.go  zstd extraction  │
│    ├── env.go      LD_LIBRARY_PATH  │
│    └── policy.go   security policy  │
├─────────────────────────────────────┤
│  runtime-<target>.tar.zst           │
│  (embedded at build time)           │
│    ├── bin/magick                    │
│    ├── lib/libMagickWand-7.*.so     │
│    ├── lib/libMagickCore-7.*.so     │
│    ├── lib/*.so (all dependencies)  │
│    ├── lib/ImageMagick-*/modules/   │
│    └── etc/ImageMagick-*/           │
└─────────────────────────────────────┘
```

## Building

The repository does not commit runtime bundles. CI builds them from source:

```sh
# Linux (builds ImageMagick from source)
bash scripts/build-runtime-linux.sh linux-amd64 internal/runtimebundle/assets/runtime-linux-amd64.tar.zst

# macOS (builds ImageMagick from source with Homebrew dependencies)
bash scripts/build-runtime-darwin.sh darwin-arm64 internal/runtimebundle/assets/runtime-darwin-arm64.tar.zst

# Then build the Go binary
CGO_ENABLED=0 go build -o dist/magickgo ./cmd/magickgo
```

CI caches the runtime tarball keyed on script content hash. First build takes
5-10 minutes; subsequent builds with unchanged scripts complete in under 3 minutes.
