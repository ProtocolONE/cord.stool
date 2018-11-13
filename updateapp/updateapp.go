package updateapp

import (
	"fmt"

	"../appargs"
)

// Start ...
func Start() bool {

	if !appargs.Init() {
		return false
	}

	UploadToFTP()

	fmt.Println("Done!")
	return true
}

func initializeSourceAndOutputPaths() {

}
