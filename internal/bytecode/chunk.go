package bytecode

type Chunk struct {
	Code      []byte
	Constants []Value
	Lines     []int
}

func (c *Chunk) addLine(line int) {
	c.Lines = append(c.Lines, line)
}

func (c *Chunk) Write(op OpCode, line int) {
	c.Code = append(c.Code, byte(op))
	c.addLine(line)
}

func (c *Chunk) WriteByte(b byte, line int) {
	c.Code = append(c.Code, b)
	c.addLine(line)
}
func (c *Chunk) WriteUint16(v uint16, line int) {
	c.Code = append(c.Code, byte(v>>8), byte(v))
	c.addLine(line)
	c.addLine(line)
}

func (c *Chunk) PatchUint16(offset int, v uint16) {
	c.Code[offset] = byte(v >> 8)
	c.Code[offset+1] = byte(v)
}

func (c *Chunk) AddConstant(v Value) int {
	c.Constants = append(c.Constants, v)
	return len(c.Constants) - 1
}
