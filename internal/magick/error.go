package magick

import (
	"errors"
	"unsafe"
)

func boolErr(ok uint, w unsafe.Pointer) error {
	if ok != falseValue {
		return nil
	}
	if w == nil || magickGetException == nil {
		return errors.New("imagemagick operation failed")
	}
	var typ ExceptionType
	msg := magickGetException(w, &typ)
	if msg == nil {
		return errors.New("imagemagick operation failed")
	}
	defer magickRelinquishMemory(unsafe.Pointer(msg))
	return errors.New(goString(msg))
}

func cstring(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}

func goString(p *byte) string {
	if p == nil {
		return ""
	}
	var n int
	for {
		if *(*byte)(unsafe.Add(unsafe.Pointer(p), n)) == 0 {
			break
		}
		n++
	}
	return string(unsafe.Slice(p, n))
}
