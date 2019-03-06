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
var _barTotal *uiprogress.Bar
var _curTitle string
var _title *string

type Args = struct {
	Url       string
	Login     string
	Password  string
	SourceDir string
	OutputDir string
	Patch     bool
	Hash      bool
	Wharf     bool
}

// Upload ...
func Upload(args Args) error {

	fmt.Println("Uploading to cord server ...")

	fullSourceDir, err := filepath.Abs(args.SourceDir)
	if err != nil {
		return err
	}

	var barCount int

	if args.Wharf {

		barCount = 10
	} else {

		fc, err := utils.FileCount(fullSourceDir)
		if err != nil {
			return err
		}
		barCount = fc + 1
	}

	uiprogress.Start()
	_barTotal = uiprogress.AddBar(barCount).AppendCompleted().PrependElapsed()
	_barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize("Total progress", 35)
	})

	api := cordapi.NewCordAPI(args.Url)
	err = api.Login(args.Login, args.Password)
	if err != nil {
		return err
	}

	if args.Wharf {

		err = uploadWharf(api, fullSourceDir, args.OutputDir)
		if err != nil {
			return err
		}

	} else {

		err = upload(api, args, fullSourceDir)
		if err != nil {
			return err
		}
	}

	_curTitle = "Finished"
	uiprogress.Stop()

	fmt.Println("Upload completed.")

	return nil
}

func upload(api *cordapi.CordAPIManager, args Args, fullSourceDir string) error {

	var err error

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

	_title = &_curTitle

	_bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(*_title, 35)
	})

	_barTotal.Incr()

	for path := range f {

		_, fn := filepath.Split(path)
		_curTitle = fmt.Sprint("Uploading file: ", fn)

		_barTotal.Incr()
		_bar.Set(0)

		err := uploadFile(api, args, path, fullSourceDir)
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

func compareHash(api *cordapi.CordAPIManager, path string, fpath string, fname string) (bool, error) {

	hash, err := utils.Md5(path)
	if err != nil {
		return false, errors.New("Hash calculating error: " + err.Error())
	}

	cmpRes, err := api.CmpHash(&models.CompareHashCmd{FilePath: fpath, FileName: fname, FileHash: hash})
	if err != nil {
		return false, err
	}

	return cmpRes.Equal, nil
}

func uploadFile(api *cordapi.CordAPIManager, args Args, path string, source string) error {

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

		r, err := compareHash(api, path, fpath, fname)
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

	err = api.Upload(&models.UploadCmd{FilePath: fpath, FileName: fname, FileData: filedata, Patch: args.Patch})
	if err != nil {
		return err
	}

	_bar.Incr()

	return nil
}
