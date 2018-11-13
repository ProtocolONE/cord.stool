package main

import (
	"fmt"

	"./appargs"
	"./updateapp"
)

func main() {

	fmt.Println("\nUpdate List Generator tool")
	if appargs.Init() {
		updateapp.Start()
	}
}
