# magick-go

`magickgo` is a pure-Go CLI wrapper around an embedded ImageMagick runtime.
The Go binary extracts `runtime-<target>.tar.zst` into the user cache, sets the
ImageMagick environment, and loads `libMagickWand` through `purego`.

Supported targets:

- `linux/amd64`
- `linux/arm64`
- `darwin/arm64`

Initial commands:

```sh
magickgo doctor --verbose
magickgo formats
magickgo identify input.png
magickgo convert input.heic output.webp
magickgo resize input.jpg output.webp --width 1200
```

The repository intentionally does not commit large runtime bundles. CI creates
`internal/runtimebundle/assets/runtime-<target>.tar.zst` before building each
target binary.

