package cord

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"cord.stool/cordapi"
	"cord.stool/service/models"
	"cord.stool/utils"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

var _bar *uiprogress.Bar

type Args = struct {
	Url       string
	Login     string
	Password  string
	SourceDir string
	OutputDir string
	Patch     bool
	Hash      bool
}

// Upload ...
func Upload(args Args) error {

	fmt.Println("Uploading to cord server ...")

	fullSourceDir, err := filepath.Abs(args.SourceDir)
	if err != nil {
		return err
	}

	fc, err := utils.FileCount(fullSourceDir)
	if err != nil {
		return err
	}

	uiprogress.Start()
	barTotal := uiprogress.AddBar(fc + 1).AppendCompleted().PrependElapsed()
	barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize("Total progress", 35)
	})

	auth, err := cordapi.Login(args.Url, args.Login, args.Password)
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

	if args.Hash {
		_bar = uiprogress.AddBar(4).AppendCompleted().PrependElapsed()
	} else {
		_bar = uiprogress.AddBar(3).AppendCompleted().PrependElapsed()
	}

	var curTitle string
	var title *string
	title = &curTitle

	_bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(*title, 35)
	})

	barTotal.Incr()

	for path := range f {

		_, fn := filepath.Split(path)
		curTitle = fmt.Sprint("Uploading file: ", fn)

		barTotal.Incr()
		_bar.Set(0)

		err := uploadFile(args, auth.Token, path, fullSourceDir)
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

func compareHash(url string, token string, path string, fpath string, fname string) (bool, error) {

	hash, err := utils.Md5(path)
	if err != nil {
		return false, errors.New("Hash calculating error: " + err.Error())
	}

	cmpRes, err := cordapi.CmpHash(url, token, &models.CompareHashCmd{FilePath: fpath, FileName: fname, FileHash: hash})
	if err != nil {
		return false, err
	}

	return cmpRes.Equal, nil
}

func uploadFile(args Args, token string, path string, source string) error {

	_bar.Incr()

	_, fname := filepath.Split(path)
	relativePath, err := filepath.Rel(source, path)
	if err != nil {
		return err
	}

	fpath := filepath.Join(args.OutputDir, relativePath)
	fpath, _ = filepath.Split(fpath)
	fpath = strings.TrimRight(fpath, "/\\")

	if args.Hash {

		r, err := compareHash(args.Url, token, path, fpath, fname)
		if err != nil {
			return errors.New("Compare hash error: " + err.Error())
		}

		_bar.Incr()

		if r {
			_bar.Incr()
			return nil // no need to upload
		}
	}

	filedata, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New("Cannot read file: " + err.Error())
	}

	_bar.Incr()

	err = cordapi.Upload(args.Url, token, &models.UploadCmd{FilePath: fpath, FileName: fname, FileData: filedata, Patch: args.Patch})
	if err != nil {
		return err
	}

	_bar.Incr()

	return nil
}
