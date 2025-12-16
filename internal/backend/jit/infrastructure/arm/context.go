package arm

import "unsafe"

// контракт между ВМ го и машинным кодом

type ContextVM struct {
	LocalsBase unsafe.Pointer
	StackBase  unsafe.Pointer
	StackSize  uint32
	_          uint32 // чтобы следующая инструкция 8 байт начиналась с адреса кратного 8
	NextIp     uint32
	DidReturn  uint32
}
