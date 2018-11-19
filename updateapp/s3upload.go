package updateapp

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"uplgen/appargs"
	"uplgen/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type enumDirCallbackS3 struct {
	sess            *session.Session
	continueOnError bool
}

// UploadToS3 ...
func UploadToS3() {

	if len(appargs.S3BucketName) == 0 || len(appargs.OutputDir) == 0 {
		return
	}

	fmt.Println("Upload to Amazon S3 Bucket ...")

	sess, err := initAWS()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var callback utils.IEnumDirCallback = enumDirCallbackS3{sess: sess, continueOnError: true}
	_, err = utils.EnumDir(appargs.OutputDir, callback)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func initAWS() (*session.Session, error) {

	var cred *credentials.Credentials
	if len(appargs.AWSCredentials) > 0 {
		cred = credentials.NewSharedCredentials(appargs.AWSCredentials, appargs.AWSProfile)
	} else {
		cred = credentials.NewStaticCredentials(appargs.AWSID, appargs.AWSKey, "")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(appargs.AWSRegion),
		Credentials: cred,
	})

	if err != nil {
		return nil, errors.New("Init AWS error: " + err.Error())
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, errors.New("Init AWS error: " + err.Error())
	}

	return sess, nil
}

func (callback enumDirCallbackS3) Process(fi os.FileInfo, path string) (bool, error) {

	if fi.IsDir() {
		return utils.EnumDir(path, callback)
	}
	return uploadFileToS3(callback.sess, fi, path)
}

func (callback enumDirCallbackS3) Error(err error, fi os.FileInfo, path string) bool {
	fmt.Println(err.Error())
	return callback.continueOnError
}

func uploadFileToS3(sess *session.Session, fi os.FileInfo, path string) (bool, error) {

	fmt.Println("Uploading file:", path)

	file, err := os.Open(path)
	if err != nil {
		return true, errors.New("Cannot open file: " + err.Error())
	}

	defer file.Close()

	fname := path[len(appargs.OutputDir)+1 : len(path)]
	fname = strings.Replace(fname, "\\", "/", -1)

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(appargs.S3BucketName),
		Key:    aws.String(fname),
		Body:   file,
	})

	if err != nil {
		return true, errors.New("Cannot upload file: " + err.Error())
	}

	return true, nil
}
