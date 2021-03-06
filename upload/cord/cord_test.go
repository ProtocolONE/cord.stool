package cord

import (
	"log"
	"testing"
)

func TestUpload(t *testing.T) {

	/*args := Args{
		SourceDir:  "..\\",
		GameID:     "ceba80b3-60de-4fbd-9ae7-7bbfece5e5e2",
		BranchName: "Test",
		Url:        "http://127.0.0.1:5001",
		Login:      "admin",
		Password:   "123456",
		Force:      true,
	}*/

	args := Args{
		SourceDir:  "D:\\Temp\\Test.new",
		GameID:     "ceba80b3-60de-4fbd-9ae7-7bbfece5e5e2",
		BranchName: "Test",
		Url:        "http://127.0.0.1:5001",
		Login:      "admin",
		Password:   "123456",
		Wharf:      true,
	}

	e := Upload(args)
	if e != nil {
		log.Panic(e)
	}
}
