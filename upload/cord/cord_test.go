package cord

import (
	"log"
	"testing"
)

func TestUpload(t *testing.T) {

	args := Args{
		SourceDir: "..\\",
		OutputDir: "test1",
		Url:       "http://127.0.0.1:5001",
		Login:     "admin",
		Password:  "123456",
	}

	e := Upload(args)
	if e != nil {
		log.Panic(e)
	}
}
