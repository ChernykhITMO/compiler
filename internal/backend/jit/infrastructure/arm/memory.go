package arm

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

type ExecuteMemory struct {
	buf      []byte // 4-байтовые arm инструкции
	usedByte int
}

func AllocateMemoryMmap(size int) (*ExecuteMemory, error) {
	if size <= 0 {
		return nil, fmt.Errorf("AllocateMemoryMmap(): size must be > 0")
	}

	page := unix.Getpagesize()
	size = ((size + page - 1) / page) * page

	data, err := unix.Mmap(-1, 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANON)
	if err != nil {
		return nil, err
	}

	return &ExecuteMemory{buf: data, usedByte: 0}, nil
}

func (mem *ExecuteMemory) FreeMemoryMmap() error {
	if mem.buf == nil {
		return nil
	}

	err := unix.Munmap(mem.buf)
	mem.buf = nil
	mem.usedByte = 0
	return err
}

func (mem *ExecuteMemory) GetBuf() []byte {
	return mem.buf
}
func (mem *ExecuteMemory) GetPtrBaseBuf() uintptr {
	if len(mem.buf) == 0 {
		return 0
	}
	return uintptr(unsafe.Pointer(&mem.buf[0]))
}
func (mem *ExecuteMemory) GetUsedByte() int {
	return mem.usedByte
}

func writeUint32(buf []byte, pos int, instr uint32) {
	buf[pos+0] = byte(instr)
	buf[pos+1] = byte(instr >> 8)
	buf[pos+2] = byte(instr >> 16)
	buf[pos+3] = byte(instr >> 24)
}

func (mem *ExecuteMemory) WriteUint32JitInstruction(instr uint32) {
	if mem.usedByte+4 > len(mem.buf) {
		panic("WriteUint32JitInstruction(): out of space bytes")
	}

	writeUint32(mem.buf, mem.usedByte, instr)
	mem.usedByte += 4
}

func (mem *ExecuteMemory) MakeReadExecute() error {
	return unix.Mprotect(mem.buf, unix.PROT_READ|unix.PROT_EXEC)
}

func (mem *ExecuteMemory) PatchUint32At(bytePos int, instr uint32) {
	writeUint32(mem.buf, bytePos, instr)
}
