package main

import (
	"fmt"
	
	"github.com/eljobe/sysos/archcode"
)

func main() {
	archName := archcode.GetArchName()
	fmt.Println("Architecture:", archName)
}
