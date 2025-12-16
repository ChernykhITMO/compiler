package bytecode

type Chunk struct {
	Code      []byte  // байткод(опкод+аргументы)
	Constants []Value // слайс констант, к которым обращается opConst
}

func (c *Chunk) Write(op OpCode) {
	c.Code = append(c.Code, byte(op))
}

func (c *Chunk) WriteByte(b byte) {
	c.Code = append(c.Code, b)
}
func (c *Chunk) WriteUint16(v uint16) {
	c.Code = append(c.Code, byte(v>>8), byte(v))
}

func (c *Chunk) PatchUint16(offset int, v uint16) {
	c.Code[offset] = byte(v >> 8)
	c.Code[offset+1] = byte(v)
}

func (c *Chunk) AddConstant(v Value) int {
	c.Constants = append(c.Constants, v)
	return len(c.Constants) - 1
}
