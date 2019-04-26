package akamai

import (
	"log"
	"testing"
)

func TestUpload(t *testing.T) {

	args := Args{
		SourceDir: "..\\",
		Hostname:  "p1-nsu.akamaihd.net",
		Keyname:   "akm-keyname",
		Key:       "akm-key",
		Code:      "akm-code",
	}

	e := Upload(args)
	if e != nil {
		log.Panic(e)
	}
}
