package ftp

import (
	"log"
	"testing"
)

func TestUpload(t *testing.T) {

	e := Upload("ftp://ftptmp:BwEHZ11C3p@fs1.gamenet.ru:21/tmp/", "d:\\temp\\test2.txt", "")
	//e := Upload("ftp://ftptmp:BwEHZ11C3p@fs1.gamenet.ru:21/ftptmp/", `..\`, "")
	if e != nil {
		log.Fatal(e)
	}
}
