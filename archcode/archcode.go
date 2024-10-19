// Package archcode defines functions for getting the architecture code for the system.
package archcode

// Architecture codes returned by GetArchCode()
const (
	ArchUnknown = iota
	ArchX86_64
	ArchX86
	ArchARM
	ArchARM64
)

// GetArchCode is implemented in assembly and C (archcode.c)
//go:noescape
func GetArchCode() int32

// GetArchName returns the architecture name as a string.
func GetArchName() string {
	code := GetArchCode()
	switch code {
	case ArchX86_64:
		return "x86_64"
	case ArchX86:
		return "x86"
	case ArchARM:
		return "ARM"
	case ArchARM64:
		return "ARM64"
	default:
		return "Unknown Architecture"
	}
}
