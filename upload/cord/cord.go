package cord

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"cord.stool/cordapi"
	utils2 "cord.stool/service/core/utils"
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
	Url        string
	Login      string
	Password   string
	GameID     string
	BranchName string
	BuildID    string
	SrcBuildID string
	SourceDir  string
	TargetDir  string
	Config     string
	Locale     string
	Platform   string
	Patch      bool
	Hash       bool
	Wharf      bool
	Force      bool
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
		barCount = fc + 1 + 3 + 1
	}

	uiprogress.Start()
	_barTotal = uiprogress.AddBar(barCount).AppendCompleted().PrependElapsed()
	_barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize("Total progress", 35)
	})

	cfg, err := utils2.ReadConfigFile(args.Config, nil)
	if err != nil {
		return errors.New("Cannot read config file: " + err.Error())
	}

	api := cordapi.NewCordAPI(args.Url)
	err = api.Login(args.Login, args.Password)
	if err != nil {
		return err
	}

	branch, build, err := createBuild(api, &args, cfg)
	if err != nil {
		return err
	}

	if args.Wharf {

		err = uploadWharf(api, args, fullSourceDir, cfg)
		if err != nil {
			return err
		}

	} else {

		err = upload(api, args, fullSourceDir, cfg)
		if err != nil {
			return err
		}
	}

	_curTitle = fmt.Sprint("Uploading config file ...")
	err = uploadFile(api, args, args.Config, "", true, cfg)
	if err != nil {
		return err
	}
	_barTotal.Incr()

	err = updateBranch(api, branch, build)
	if err != nil {
		return err
	}

	_curTitle = "Finished"
	uiprogress.Stop()

	fmt.Println("Upload completed.")

	return nil
}

func createBuild(api *cordapi.CordAPIManager, args *Args, cfg *models.Config) (*models.Branch, *models.Build, error) {

	_curTitle = fmt.Sprint("Creating build ...")

	branch, err := api.GetBranch("", args.BranchName, args.GameID)
	if err != nil {

		if !args.Force {
			return nil, nil, err
		}

		_curTitle = fmt.Sprint("Creating branch ...")
		branch, err = api.CreateBranch(&models.Branch{"", args.BranchName, args.GameID, "", false, time.Time{}, time.Time{}})
		if err != nil {
			return nil, nil, err
		}
	}

	_barTotal.Incr()
	_curTitle = fmt.Sprint("Creating build ...")

	build, err := api.CreateBuild(&models.Build{"", branch.ID, time.Time{}, cfg.Application.Platform /*, "", "", "", "", ""*/})
	if err != nil {
		return nil, nil, err
	}
	args.BuildID = build.ID

	if args.Wharf {
		build, err := api.GetLiveBuild(args.GameID, args.BranchName)
		if err != nil {
			return nil, nil, err
		}
		args.SrcBuildID = build.ID
	}

	_barTotal.Incr()
	return branch, build, nil
}

func updateBranch(api *cordapi.CordAPIManager, branch *models.Branch, build *models.Build) error {

	_curTitle = fmt.Sprint("Updating branch ...")

	branch.LiveBuild = build.ID
	err := api.UpdateBranch(branch)
	if err != nil {
		return err
	}

	_barTotal.Incr()
	return nil
}

func upload(api *cordapi.CordAPIManager, args Args, fullSourceDir string, cfg *models.Config) error {

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

		err := uploadFile(api, args, path, fullSourceDir, false, cfg)
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

func compareHash(api *cordapi.CordAPIManager, path string, buildid string, fpath string, fname string) (bool, error) {

	hash, err := utils.Md5(path)
	if err != nil {
		return false, errors.New("Hash calculating error: " + err.Error())
	}

	cmpRes, err := api.CmpHash(&models.CompareHashCmd{BuildID: buildid, FilePath: fpath, FileName: fname, FileHash: hash})
	if err != nil {
		return false, err
	}

	return cmpRes.Equal, nil
}

func uploadFile(api *cordapi.CordAPIManager, args Args, path string, source string, config bool, cfg *models.Config) error {

	_bar.Incr()

	_, fname := filepath.Split(path)

	fpath := ""
	if !config {

		relativePath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		fpath = relativePath
		fpath, _ = filepath.Split(fpath)
		fpath = strings.TrimRight(fpath, "/\\")
	}

	if args.Hash {

		r, err := compareHash(api, path, args.BuildID, fpath, fname)
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

	err = api.Upload(&models.UploadCmd{BuildID: args.BuildID, FilePath: fpath, FileName: fname, FileData: filedata, Patch: args.Patch, Config: config, Platform: cfg.Application.Platform})
	if err != nil {
		return err
	}

	_bar.Incr()

	return nil
}
