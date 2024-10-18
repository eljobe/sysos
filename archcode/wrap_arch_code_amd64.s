// +build amd64

#include "textflag.h"

TEXT ·GetArchCode(SB), NOSPLIT, $0-4
	// Call the external C function
	CALL GetArchCode(SB)
	// Move the return value from AX to the Go return location
	MOVL AX, ret+0(FP)
	RET
