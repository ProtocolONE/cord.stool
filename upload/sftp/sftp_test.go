package sftp

import (
	"log"
	"testing"
)

func TestUpload(t *testing.T) {

	e := Upload("sftp://user:password@ams.origin.gcdn.co:2200/htdocs/test", `..\`)
	if e != nil {
		log.Panic(e)
	}
}
