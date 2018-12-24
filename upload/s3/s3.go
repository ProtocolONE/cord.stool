package s3

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"cord.stool/utils"
	"github.com/gosuri/uiprogress"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var _bar *uiprogress.Bar

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

	fmt.Println("Uploading to Amazon S3 Bucket ...")
	
	fullSourceDir, err := filepath.Abs(args.SourceDir)
	if err != nil {
		return err
	}

	fc, err := utils.FileCount(fullSourceDir)
	if err != nil {
		return err
	}

	uiprogress.Start()
	barTotal := uiprogress.AddBar(fc + 1 ).AppendCompleted().PrependElapsed()
	barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return "Total progress"
	})

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

	_bar = uiprogress.AddBar(3).AppendCompleted().PrependElapsed()

	var curTitle string
	var title *string
	title = &curTitle

	_bar.PrependFunc(func(b *uiprogress.Bar) string {
		return *title
	})

	barTotal.Incr();

	for path := range f {

		_, fn := filepath.Split(path)
		curTitle = fmt.Sprint("Uploading file: ", fn)

		barTotal.Incr();
		_bar.Set(0);

		err := uploadFile(sess, args.OutputDir, path, fullSourceDir, args.S3Bucket)
		if err != nil {
			return err
		}
	}

	err = <-e
	if err != nil {
		return err
	}

	curTitle = "Finished"
	uiprogress.Stop()

	fmt.Println("Upload completed.")
	
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

	_bar.Incr();

	file, err := os.Open(path)
	if err != nil {
		return errors.New("Cannot open file: " + err.Error())
	}
	defer file.Close()

	_bar.Incr();

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

	_bar.Incr();

	return nil
}
