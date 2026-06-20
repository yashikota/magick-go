package magick

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"unsafe"
)

type ConvertOptions struct {
	Quality    int
	Strip      bool
	AutoOrient bool
	Format     string
}

type ImageInfo struct {
	Path   string `json:"path"`
	Width  uint64 `json:"width"`
	Height uint64 `json:"height"`
	Depth  uint64 `json:"depth"`
	Format string `json:"format"`
}

func NewWand() (*Wand, error) {
	ptr := newMagickWand()
	if ptr == nil {
		return nil, fmt.Errorf("NewMagickWand returned nil")
	}
	return &Wand{ptr: ptr}, nil
}

func (w *Wand) Close() {
	if w != nil && w.ptr != nil {
		w.ptr = destroyMagickWand(w.ptr)
	}
}

func Read(path string) (*Wand, error) {
	w, err := NewWand()
	if err != nil {
		return nil, err
	}
	if err := boolErr(magickReadImage(w.ptr, &cstring(path)[0]), w.ptr); err != nil {
		w.Close()
		return nil, err
	}
	return w, nil
}

func Identify(path string) (*ImageInfo, error) {
	w, err := Read(path)
	if err != nil {
		return nil, err
	}
	defer w.Close()
	return w.Info(path), nil
}

func Convert(input, output string, opts ConvertOptions) error {
	w, err := Read(input)
	if err != nil {
		return err
	}
	defer w.Close()
	if err := w.applyOptions(output, opts); err != nil {
		return err
	}
	return boolErr(magickWriteImages(w.ptr, &cstring(output)[0], trueValue), w.ptr)
}

func Resize(input, output string, width uint64, opts ConvertOptions) error {
	w, err := Read(input)
	if err != nil {
		return err
	}
	defer w.Close()
	if width > 0 {
		ow := uint64(magickGetImageWidth(w.ptr))
		oh := uint64(magickGetImageHeight(w.ptr))
		if ow == 0 || oh == 0 {
			return fmt.Errorf("cannot determine image size")
		}
		height := uint64(math.Round(float64(oh) * float64(width) / float64(ow)))
		if height == 0 {
			height = 1
		}
		if err := boolErr(magickResizeImage(w.ptr, uintptr(width), uintptr(height), lanczosFilter, 1.0), w.ptr); err != nil {
			return err
		}
	}
	if err := w.applyOptions(output, opts); err != nil {
		return err
	}
	return boolErr(magickWriteImages(w.ptr, &cstring(output)[0], trueValue), w.ptr)
}

func (w *Wand) Info(path string) *ImageInfo {
	format := ""
	if p := magickGetImageFormat(w.ptr); p != nil {
		format = goString(p)
		magickRelinquishMemory(unsafe.Pointer(p))
	}
	return &ImageInfo{
		Path:   path,
		Width:  uint64(magickGetImageWidth(w.ptr)),
		Height: uint64(magickGetImageHeight(w.ptr)),
		Depth:  uint64(magickGetImageDepth(w.ptr)),
		Format: format,
	}
}

func (w *Wand) applyOptions(output string, opts ConvertOptions) error {
	if opts.AutoOrient {
		if err := boolErr(magickAutoOrientImage(w.ptr), w.ptr); err != nil {
			return err
		}
	}
	if opts.Strip {
		if err := boolErr(magickStripImage(w.ptr), w.ptr); err != nil {
			return err
		}
	}
	if opts.Quality > 0 {
		if err := boolErr(magickSetImageCompressionQuality(w.ptr, uintptr(opts.Quality)), w.ptr); err != nil {
			return err
		}
	}
	format := opts.Format
	if format == "" {
		format = strings.TrimPrefix(filepath.Ext(output), ".")
	}
	if format != "" {
		format = strings.ToUpper(format)
		if err := boolErr(magickSetImageFormat(w.ptr, &cstring(format)[0]), w.ptr); err != nil {
			return err
		}
	}
	return nil
}
