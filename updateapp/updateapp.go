package updateapp

import (
	"fmt"

	"../appargs"
)

// Start ...
func Start() int {

	if !appargs.Init() {
		return 1;
	}

	fmt.Println(appargs.SourceDir)
	return 0
}

func initializeSourceAndOutputPaths() {
	
}