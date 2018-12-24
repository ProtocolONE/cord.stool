package ftp

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func UploadToFTPError(t *testing.T) {
	e := Upload("ftp://ftpuser:ftppass@ftp.protocol.local:21/cordtest/", `..\`)

	if e != nil {
		log.Fatal(e)
	}
	fmt.Println("test")
	assert.True(t, true)
	assert.True(t, e == nil)
}
