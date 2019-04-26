package ftp

import (
	"log"
	"testing"
)

func TestUpload(t *testing.T) {

	e := Upload("ftp://ftpuser:ftppass@ftp.protocol.local:21/cordtest/", `..\`)
	if e != nil {
		log.Fatal(e)
	}
}
