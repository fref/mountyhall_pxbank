package main

import (
	"fmt"
	"os"
)

const CANNOT_RESTORE_BANK = -1

func main() {
	var bank PxBank
	err := bank.Reload()
	if nil != err {
		fmt.Println("Could not reload PxBank storage: ", err)
		os.Exit(CANNOT_RESTORE_BANK)
	}

	bank.LoopMenu(DisplayMainMenu, HandleMainChoice)
}
