// +build amd64

#include "textflag.h"

TEXT ·GetArchCode(SB), NOSPLIT, $0-4
	CALL GetArchCode(SB)
	MOVL AX, ret+0(FP)
	RET
