package update

import (
	//"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"cord.stool/cordapi"
	"cord.stool/service/core/utils"
	"cord.stool/service/models"
	"cord.stool/upload/cord"
	utils2 "cord.stool/utils"

	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"github.com/itchio/wharf/state"
)

var _bar *uiprogress.Bar
var _barTotal *uiprogress.Bar
var _curTitle string
var _totalTitle string

func newStateConsumer() *state.Consumer {
	return &state.Consumer{
		OnProgress:       Progress,
		OnProgressLabel:  ProgressLabel,
		OnPauseProgress:  PauseProgress,
		OnResumeProgress: ResumeProgress,
		OnMessage:        Logl,
	}
}

func ProgressLabel(label string) {

	//_, fn := filepath.Split(label)
	//_curTitle = fn
}

func PauseProgress() {
}

func ResumeProgress() {
}

func Progress(alpha float64) {

	_bar.Set(int(100 * alpha))
	//_barTotal.Set(int(5*alpha) + (_barTotal.Total - 7))
}

func Logl(level string, msg string) {
}

func Update(args cord.Args) error {

	fmt.Println("Updating game ...")

	uiprogress.Start()
	_barTotal = uiprogress.AddBar(2).AppendCompleted().PrependElapsed()
	_barTotal.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(_totalTitle, 35)
	})
	_totalTitle = "Total progress"
	_barTotal.Total = 6

	_bar = uiprogress.AddBar(2).AppendCompleted().PrependElapsed()
	_bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(_curTitle, 35)
	})

	needInstall := false
	usePatch := false
	gameVer := getGameVersion(args.TargetDir, args.Platform)
	if gameVer != "" {
		usePatch = true
	}

	contentPath := path.Join(args.TargetDir, "content")
	patchFile := path.Join(args.TargetDir, "patch_for_"+gameVer+"_"+args.Platform+".bin")
	var manifest *models.ConfigManifest
	var err error

	if _, err := os.Stat(patchFile); !os.IsNotExist(err) {

		defer os.Remove(patchFile)

		_bar.Total = 100
		_curTitle = "Applying patch ..."

		err := utils2.ApplyPatchFile(contentPath, contentPath, patchFile, newStateConsumer())
		if err == nil {

			_barTotal.Incr()
			_barTotal.Incr()

			data, err := ioutil.ReadFile(path.Join(args.TargetDir, "config.json"))
			if err != nil {
				return err
			}

			_barTotal.Incr()

			manifest, err = getManifest(data, args.Platform)
			if err != nil {
				return err
			}

			_barTotal.Incr()
			_barTotal.Incr()

		} else {

			manifest, needInstall, err = doUpdate(args, usePatch, gameVer)
			if err != nil {
				return err
			}
		}

	} else {

		manifest, needInstall, err = doUpdate(args, usePatch, gameVer)
		if err != nil {
			return err
		}
	}

	if needInstall {
		err = doInstall(args, manifest)
		if err != nil {
			return err
		}
	} else {
		_barTotal.Incr()
	}

	_curTitle = "Finished"
	uiprogress.Stop()

	fmt.Println("Update completed.")

	return nil
}

func doUpdate(args cord.Args, usePatch bool, gameVer string) (*models.ConfigManifest, bool, error) {

	_curTitle = "Getting update info"
	api := cordapi.NewCordAPI(args.Url)
	err := api.Login(args.Login, args.Password)
	if err != nil {
		return nil, false, err
	}

	_barTotal.Incr()
	_bar.Incr()

	contentPath := path.Join(args.TargetDir, "content")
	torrentFile := path.Join(args.TargetDir, "torrent.torrent")
	var info *models.UpdateInfo
	needInstall := false

	if usePatch {

		info, err = api.GetUpdatePatch(args.GameID, args.BranchName, args.Locale, args.Platform, gameVer)
		if err != nil {

			usePatch = false // try to download whole build

		} else {

			_curTitle = "Checking"
			err := VerifyTorrentFile(torrentFile, args.TargetDir, _bar)
			if err != nil {
				usePatch = false // try to download whole build
			}
		}
	}

	if !usePatch {

		_barTotal.Incr()

		info, err = api.GetUpdateInfo(args.GameID, args.BranchName, args.Locale, args.Platform, gameVer)
		if err != nil {
			return nil, false, err
		}
	}

	_barTotal.Incr()
	_bar.Incr()

	stats := NewDownloadStatistics("update")

	if gameVer == info.Version {

		_curTitle = "Checking"

		err = VerifyTorrentFile(torrentFile, args.TargetDir, _bar)
		if err != nil {

			_curTitle = "Downloading"
			err = StartDownloadFile(torrentFile, args.TargetDir, _bar, stats)
			if err != nil {
				return nil, false, err
			}
		}

	} else {

		_curTitle = "Downloading"

		if usePatch {
			err = StartDownload(info.TorrentPatchData, args.TargetDir, _bar, stats)
		} else {
			err = StartDownload(info.TorrentData, args.TargetDir, _bar, stats)
			needInstall = true
		}

		if err != nil {
			return nil, false, err
		}
	}

	_barTotal.Incr()

	if usePatch {

		patchFile := path.Join(args.TargetDir, "patch_for_"+gameVer+"_"+args.Platform+".bin")
		defer os.Remove(patchFile)

		_bar.Total = 100
		_curTitle = "Applying patch ..."

		err = utils2.ApplyPatchFile(contentPath, contentPath, patchFile, newStateConsumer())
		if err != nil {

			_curTitle = "Downloading"
			err = StartDownload(info.TorrentData, args.TargetDir, _bar, stats)
			if err != nil {
				return nil, false, err
			}
		}

		_barTotal.Incr()
	}

	_bar.Set(0)
	_bar.Total = 3
	_curTitle = "Preparing ..."

	err = ioutil.WriteFile(torrentFile, info.TorrentData, 0777)
	if err != nil {
		return nil, false, err
	}

	_bar.Incr()

	err = ioutil.WriteFile(path.Join(args.TargetDir, "config.json"), info.ConfigData, 0777)
	if err != nil {
		return nil, false, err
	}

	_bar.Incr()

	manifest, err := getManifest(info.ConfigData, args.Platform)
	if err != nil {
		return nil, false, err
	}

	_bar.Incr()
	_barTotal.Incr()

	return manifest, needInstall, nil
}

func doInstall(args cord.Args, manifest *models.ConfigManifest) error {

	_bar.Set(0)
	_bar.Total = 3
	_curTitle = "Installing game ..."

	err := downloadRedist(manifest)
	if err != nil {
		return err
	}

	_bar.Incr()

	err = install(args.TargetDir, manifest)
	if err != nil {
		return err
	}

	_bar.Incr()

	err = utils2.AddRegKeys(manifest)
	if err != nil {
		return err
	}

	_bar.Incr()
	_barTotal.Incr()

	return nil
}

func getManifest(data []byte, platform string) (*models.ConfigManifest, error) {

	cfg, err := utils.ReadConfigData(data, nil)
	if err != nil {
		return nil, err
	}

	var manifest *models.ConfigManifest
	manifest = nil
	for _, m := range cfg.Application.Manifests {

		if m.Platform == platform {
			manifest = &m
			break
		}
	}

	if manifest == nil {
		return nil, fmt.Errorf("Manifests for specified platform (%s) is not found", platform)
	}

	return manifest, nil
}

func getGameVersion(target string, platform string) string {

	res, err := utils2.IsDirectoryEmpty(path.Join(target, "content"))
	if res != false || err != nil {
		return ""
	}

	cfg, err := utils.ReadConfigFile(path.Join(target, "config.json"), nil)
	if err != nil {
		return ""
	}

	for _, m := range cfg.Application.Manifests {

		if m.Platform == platform {
			return m.Version
		}
	}

	return ""
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

func install(targetDir string, manifest *models.ConfigManifest) error {

	for _, scr := range manifest.InstallScripts {

		fpath := filepath.Join(targetDir, "content", scr.Executable)
		err := utils2.RunCommand(scr.RequiresAdmin, fpath, scr.Arguments...)
		if err != nil {
			return err
		}

		completion := false

		for !completion {

			completion, err = utils2.CheckCompletion(scr.CompletionRegistryKey)
			if err != nil {
				return err
			}
			time.Sleep(1 * time.Second)
		}
	}

	_bar.Incr()

	return nil
}

func downloadRedist(manifest *models.ConfigManifest) error {

	for _, r := range manifest.Redistributables {

		url := r

		if url != "" {

			tmpDir, err := ioutil.TempDir(os.TempDir(), "p1-")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tmpDir)

			_, fname := filepath.Split(url)
			tmpfn := filepath.Join(tmpDir, fname)

			err = utils2.DownloadFile(tmpfn, url)
			if err != nil {
				return err
			}

			exec.Command(tmpfn).Run()
		}
	}

	return nil
}
