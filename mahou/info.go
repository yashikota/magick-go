package mahou

import (
	"sort"
	"unsafe"
)

type Diagnostics struct {
	Version      string            `json:"version"`
	QuantumDepth string            `json:"quantum_depth"`
	HDRI         string            `json:"hdri"`
	Configure    map[string]string `json:"configure"`
	Formats      []string          `json:"formats"`
	Support      map[string]bool   `json:"support"`
}

func Version() string {
	var version uintptr
	p := magickGetVersion(&version)
	return goString(p)
}

func ConfigureOption(name string) string {
	p := magickQueryConfigureOption(&cstring(name)[0])
	if p == nil {
		return ""
	}
	defer magickRelinquishMemory(unsafe.Pointer(p))
	return goString(p)
}

func Formats() []string {
	var n uint64
	p := magickQueryFormats(&cstring("*")[0], &n)
	if p == nil || n == 0 {
		return nil
	}
	defer magickRelinquishMemory(unsafe.Pointer(p))
	formats := make([]string, 0, n)
	slice := unsafe.Slice(p, n)
	for _, item := range slice {
		if item != nil {
			formats = append(formats, goString(item))
		}
	}
	sort.Strings(formats)
	return formats
}

func DiagnosticsInfo() Diagnostics {
	formats := Formats()
	support := make(map[string]bool)
	set := make(map[string]struct{}, len(formats))
	for _, f := range formats {
		set[f] = struct{}{}
	}
	for _, name := range []string{"JPEG", "PNG", "WEBP", "TIFF", "HEIC", "JXL", "SVG", "PDF"} {
		_, ok := set[name]
		support[name] = ok
	}
	return Diagnostics{
		Version:      Version(),
		QuantumDepth: ConfigureOption("QuantumDepth"),
		HDRI:         ConfigureOption("HDRI"),
		Configure: map[string]string{
			"CONFIGURE_PATH": ConfigureOption("CONFIGURE_PATH"),
			"DELEGATES":      ConfigureOption("DELEGATES"),
			"FEATURES":       ConfigureOption("FEATURES"),
		},
		Formats: formats,
		Support: support,
	}
}
