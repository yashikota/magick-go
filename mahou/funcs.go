package mahou

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	magickWandGenesis                func()
	magickGetVersion                 func(*uintptr) *byte
	magickQueryConfigureOption       func(*byte) *byte
	magickQueryFormats               func(*byte, *uint64) **byte
	magickRelinquishMemory           func(unsafe.Pointer) unsafe.Pointer
	newMagickWand                    func() unsafe.Pointer
	destroyMagickWand                func(unsafe.Pointer) unsafe.Pointer
	magickReadImage                  func(unsafe.Pointer, *byte) uint
	magickWriteImages                func(unsafe.Pointer, *byte, uint) uint
	magickGetImageWidth              func(unsafe.Pointer) uintptr
	magickGetImageHeight             func(unsafe.Pointer) uintptr
	magickGetImageDepth              func(unsafe.Pointer) uintptr
	magickGetImageFormat             func(unsafe.Pointer) *byte
	magickSetImageFormat             func(unsafe.Pointer, *byte) uint
	magickSetImageCompressionQuality func(unsafe.Pointer, uintptr) uint
	magickStripImage                 func(unsafe.Pointer) uint
	magickAutoOrientImage            func(unsafe.Pointer) uint
	magickResizeImage                func(unsafe.Pointer, uintptr, uintptr, uintptr, float64) uint
	magickGetException               func(unsafe.Pointer, *ExceptionType) *byte
)

func register(handle uintptr) error {
	purego.RegisterLibFunc(&magickWandGenesis, handle, "MagickWandGenesis")
	purego.RegisterLibFunc(&magickGetVersion, handle, "MagickGetVersion")
	purego.RegisterLibFunc(&magickQueryConfigureOption, handle, "MagickQueryConfigureOption")
	purego.RegisterLibFunc(&magickQueryFormats, handle, "MagickQueryFormats")
	purego.RegisterLibFunc(&magickRelinquishMemory, handle, "MagickRelinquishMemory")
	purego.RegisterLibFunc(&newMagickWand, handle, "NewMagickWand")
	purego.RegisterLibFunc(&destroyMagickWand, handle, "DestroyMagickWand")
	purego.RegisterLibFunc(&magickReadImage, handle, "MagickReadImage")
	purego.RegisterLibFunc(&magickWriteImages, handle, "MagickWriteImages")
	purego.RegisterLibFunc(&magickGetImageWidth, handle, "MagickGetImageWidth")
	purego.RegisterLibFunc(&magickGetImageHeight, handle, "MagickGetImageHeight")
	purego.RegisterLibFunc(&magickGetImageDepth, handle, "MagickGetImageDepth")
	purego.RegisterLibFunc(&magickGetImageFormat, handle, "MagickGetImageFormat")
	purego.RegisterLibFunc(&magickSetImageFormat, handle, "MagickSetImageFormat")
	purego.RegisterLibFunc(&magickSetImageCompressionQuality, handle, "MagickSetImageCompressionQuality")
	purego.RegisterLibFunc(&magickStripImage, handle, "MagickStripImage")
	purego.RegisterLibFunc(&magickAutoOrientImage, handle, "MagickAutoOrientImage")
	purego.RegisterLibFunc(&magickResizeImage, handle, "MagickResizeImage")
	purego.RegisterLibFunc(&magickGetException, handle, "MagickGetException")
	return nil
}
