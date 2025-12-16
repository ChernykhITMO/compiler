#include "textflag.h"

TEXT ·callJitEntry(SB), NOSPLIT, $0-24
	MOVD addressCode+0(FP), R2
	MOVD ctx+8(FP), R0
	BL  (R2)
	MOVW R0, ret+16(FP)
	RET

