package s3

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"cord.stool/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Args = struct {
	SourceDir string
	OutputDir string
	AWSRegion string
	AWSCredentials string
	AWSProfile string
	AWSID string
	AWSKey string
	AWSToken string
	S3Bucket string
}

type enumDirCallbackS3 struct {
	sess            *session.Session
	continueOnError bool
}

// Upload ...
func Upload(args Args) error {

	fmt.Println("Upload to Amazon S3 Bucket ...")

	fullSourceDir, err := filepath.Abs(args.SourceDir)
	if err != nil {
		return err
	}

	sess, err := initAWS(args)
	if err != nil {
		return err
	}

	stopCh := make(chan struct{})
	defer func() {
		select {
		case stopCh <- struct{}{}:
		default:
		}

		close(stopCh)
	}()

	f, e := utils.EnumFilesRecursive(fullSourceDir, stopCh)
	
	for path := range f {

		err := uploadFile(sess, args.OutputDir, path, fullSourceDir, args.S3Bucket)
		if err != nil {
			return err
		}
	}

	err = <-e
	if err != nil {
		return err
	}

	return nil
}

func initAWS(args Args) (*session.Session, error) {

	var cred *credentials.Credentials
	if len(args.AWSCredentials) > 0 {
		cred = credentials.NewSharedCredentials(args.AWSCredentials, args.AWSProfile)
	} else {
		cred = credentials.NewStaticCredentials(args.AWSID, args.AWSKey, args.AWSToken)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(args.AWSRegion),
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

func uploadFile(sess *session.Session, root string, path string, source string, bucket string) error {

	fmt.Println("Uploading file:", path)

	file, err := os.Open(path)
	if err != nil {
		return errors.New("Cannot open file: " + err.Error())
	}

	defer file.Close()

	fname := path[len(source)+1 : len(path)]
	fname = filepath.Join(root, fname)
	fname = strings.Replace(fname, "\\", "/", -1)

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fname),
		Body:   file,
	})

	if err != nil {
		return errors.New("Cannot upload file: " + err.Error())
	}

	return nil
}
