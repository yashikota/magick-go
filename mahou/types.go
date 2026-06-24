package mahou

import "unsafe"

type Wand struct {
	ptr unsafe.Pointer
}

type ExceptionType int32

const (
	falseValue uint = 0
	trueValue  uint = 1

	lanczosFilter uintptr = 22
)
