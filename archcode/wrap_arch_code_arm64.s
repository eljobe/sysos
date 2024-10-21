// +build arm64

#include "textflag.h"

TEXT ·GetArchCode(SB), NOSPLIT, $0
    BL GetArchCode(SB)
    MOVW R0, ret+0(FP)
    RET
