package update

import (
	//"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"cord.stool/cordapi"
	"cord.stool/service/core/utils"
	"cord.stool/service/models"
	"cord.stool/upload/cord"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

var _bar *uiprogress.Bar
var _barTotal *uiprogress.Bar
var _curTitle string
var _totalTitle string

func Update(args cord.Args) error {

	fmt.Println("Updating game ...")

	uiprogress.Start()
	_barTotal = uiprogress.AddBar(2).AppendCompleted().PrependElapsed()
	_barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(_totalTitle, 35)
	})
	_totalTitle = "Getting update info"

	api := cordapi.NewCordAPI(args.Url)
	err := api.Login(args.Login, args.Password)
	if err != nil {
		return err
	}

	_barTotal.Incr()

	info, err := api.GetUpdateInfo(args.GameID, args.BranchName, args.Locale, args.Platform)
	if err != nil {
		return err
	}

	_barTotal.Incr()

	_barTotal.Total = len(info.Files) + 1
	_totalTitle = "Total progress"

	_bar = uiprogress.AddBar(3).AppendCompleted().PrependElapsed()
	_bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(_curTitle, 35)
	})

	_, fn := filepath.Split(info.Config)
	_curTitle = fmt.Sprint("Downloading file: ", fn)
	err = downloadAndSave(api, info.BuildID, info.Config, args.TargetDir, args.Platform, nil)
	if err != nil {
		return err
	}

	_barTotal.Incr()

	fpath, fname := filepath.Split(info.Config)
	fpath = path.Join(args.TargetDir, fpath)
	fpath = path.Join(fpath, fname)

	cfg, err := utils.ReadConfigFile(fpath, nil)
	if err != nil {
		return err
	}

	var manifest *models.ConfigManifest
	manifest = nil

	for _, m := range cfg.Application.Manifests {

		if m.Platform == args.Platform {
			manifest = &m
			break
		}
	}

	for _, f := range info.Files {

		_, fn := filepath.Split(f)
		_curTitle = fmt.Sprint("Downloading file: ", fn)
		err = downloadAndSave(api, info.BuildID, f, args.TargetDir, args.Platform, manifest)
		if err != nil {
			return err
		}

		_barTotal.Incr()
	}

	_curTitle = "Finished"
	uiprogress.Stop()

	fmt.Println("Update completed.")

	return nil
}

func downloadAndSave(api *cordapi.CordAPIManager, buildID string, source string, target string, platform string, manifest *models.ConfigManifest) error {

	_bar.Set(0)

	data, err := api.Download(buildID, source, platform)
	if err != nil {
		return err
	}

	_bar.Set(1)

	userData := false

	if manifest != nil {

		att := getAttributes(source, manifest)
		for _, a := range att {
			
			if a == "user_data" {
				userData = true
			}
		}

		source = mapPath(source, manifest)
	}

	fpath, fname := filepath.Split(source)
	fpath = path.Join(target, fpath)

	err = os.MkdirAll(fpath, 0777)
	if err != nil {
		return err
	}

	_bar.Set(2)

	fpath = path.Join(fpath, fname)

	 _, err = os.Stat(fpath)
	if os.IsNotExist(err) || !userData {

		err = ioutil.WriteFile(fpath, data.FileData, 0777)
		if err != nil {
			return err
		}
	}

	_bar.Set(3)
	return nil
}

func mapPath(fpath string, manifest *models.ConfigManifest) string {

	fpath = filepath.ToSlash(fpath)

	for _, m := range manifest.FileRules.Mappings {

		localPath := strings.TrimLeft(m.LocalPath, ".")
		localPath = strings.TrimLeft(m.LocalPath, "/")
		localPath = filepath.Join("content", localPath) + "/"
		localPath = filepath.ToSlash(localPath)

		installPath := strings.TrimLeft(m.InstallPath, ".")
		installPath = strings.TrimLeft(m.InstallPath, "/")
		installPath = filepath.ToSlash(installPath)

		index := strings.Index(strings.ToLower(fpath), strings.ToLower(localPath))
		if index == 0 {

			fpath = filepath.Join("content", installPath, fpath[len(localPath):])
			break

		} else {

			localPath = strings.TrimRight(localPath, "/")
			if len(fpath) == len(localPath) && strings.Index(strings.ToLower(fpath), strings.ToLower(localPath)) == 0 {

				fpath = filepath.Join("content", installPath, fpath[len(localPath):])
				break
			}
		}
	}

	return fpath
}

func getAttributes(fpath string, manifest *models.ConfigManifest) []string {

	fpath = filepath.ToSlash(fpath)
	
	for _, p := range manifest.FileRules.Properties {

		prop := strings.TrimLeft(p.InstallPath, ".")
		prop = strings.TrimLeft(p.InstallPath, "/")
		prop = filepath.Join("content", prop)
		prop = filepath.ToSlash(prop)

		match, err := filepath.Match(prop, fpath)
		if match && err == nil {
			return p.Attributes
		}
	}

	return nil
}
