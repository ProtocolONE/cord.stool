package s3

import (
	"log"
	"testing"
)

func TestUpload(t *testing.T) {

	args := Args{
		SourceDir:   "..\\",
		Region:      "eu-west-3",
		Credentials: "aws-credentials",
		S3Bucket:    "protocol-one-test",
	}

	e := Upload(args)
	if e != nil {
		log.Panic(e)
	}
}
