package arm

func callJitEntry(addressCode uintptr, ctx *ContextVM) (ret uint32)

func CallJitBlock(addressCode uintptr, ctx *ContextVM) uint32 {
	return callJitEntry(addressCode, ctx)
}
