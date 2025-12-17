package arm

import "unsafe"

type ContextVM struct {
	LocalsBase unsafe.Pointer
	StackBase  unsafe.Pointer
	StackSize  uint32
	_          uint32 // чтобы следующая инструкция 8 байт начиналась с адреса кратного 8
	DidReturn  uint32
}
