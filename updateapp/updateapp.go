package updateapp

import (
	"fmt"

	"cord.stool/appargs"
)

// Start ...
func Start() bool {

	if !appargs.Init() {
		return false
	}

	if len(appargs.FTPUrl) > 0 {
		UploadToFTP()
	}

	if len(appargs.S3BucketName) > 0 {
		UploadToS3()
	}

	fmt.Println("Done!")
	return true
}

func initializeSourceAndOutputPaths() {

}
