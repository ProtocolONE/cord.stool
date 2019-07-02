package update

import (
	//"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
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

const (
	noNeed      = 0
	redistNeed  = 1
	instralNeed = 2
	registyNeed = 4
	allNeed     = -1
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
}

func PauseProgress() {
}

func ResumeProgress() {
}

func Progress(alpha float64) {

	_bar.Set(int(100 * alpha))
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

	needInstall := noNeed
	usePatch := false
	gameVer := getGameVersion(args.TargetDir, args.Platform)
	if gameVer != "" {
		usePatch = true
	}

	var manifestOld *models.ConfigManifest

	if args.Recheck {

		needInstall = allNeed

	} else {

		data, err := ioutil.ReadFile(path.Join(args.TargetDir, "config.json"))
		if err != nil {
			needInstall = allNeed
		}

		manifestOld, err = getManifest(data, args.Platform)
		if err != nil {
			needInstall = allNeed
		}
	}

	contentPath := path.Join(args.TargetDir, "content")
	patchFile := path.Join(args.TargetDir, "patch_for_"+gameVer+"_"+args.Platform+".bin")
	var manifest *models.ConfigManifest

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

		}

		info, err := getUpdateInfo(args, &usePatch, gameVer, true)
		if info != nil {
			manifest, err = doUpdate(args, usePatch, gameVer, info)
			if err != nil {
				return err
			}
		}

	} else {

		manifest, err = doUpdate(args, usePatch, gameVer, nil)
		if err != nil {
			return err
		}
	}

	if needInstall == noNeed {
		needInstall = isNeedInstall(manifestOld, manifest)
	}

	if needInstall != noNeed {
		err := doInstall(args, manifest, needInstall)
		if err != nil {
			return err
		}
	}

	_barTotal.Incr()

	_curTitle = "Finished"
	uiprogress.Stop()

	fmt.Println("Update completed.")

	return nil
}

func doUpdate(args cord.Args, usePatch bool, gameVer string, info *models.UpdateInfo) (*models.ConfigManifest, error) {

	_barTotal.Incr()
	_bar.Incr()

	contentPath := path.Join(args.TargetDir, "content")
	torrentFile := path.Join(args.TargetDir, "torrent.torrent")
	var err error

	if info == nil {

		info, err = getUpdateInfo(args, &usePatch, gameVer, false)
		if err != nil {
			return nil, err
		}
	}

	_barTotal.Incr()
	_bar.Incr()

	stats := NewDownloadStatistics("update")

	if args.Recheck {

		_curTitle = "Downloading"
		err = StartDownloadFile(torrentFile, args.TargetDir, _bar, stats)
		if err != nil {
			return nil, err
		}

	} else if gameVer == info.Version {

		_curTitle = "Checking"

		err = VerifyTorrentFile(torrentFile, args.TargetDir, _bar)
		if err != nil {

			_curTitle = "Downloading"
			err = StartDownloadFile(torrentFile, args.TargetDir, _bar, stats)
			if err != nil {
				return nil, err
			}
		}

	} else {

		_curTitle = "Downloading"

		if usePatch {
			err = StartDownload(info.TorrentPatchData, args.TargetDir, _bar, stats)
		} else {
			err = StartDownload(info.TorrentData, args.TargetDir, _bar, stats)
		}

		if err != nil {
			return nil, err
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
				return nil, err
			}
		}

		_barTotal.Incr()
	}

	_bar.Set(0)
	_bar.Total = 3
	_curTitle = "Preparing ..."

	err = ioutil.WriteFile(torrentFile, info.TorrentData, 0777)
	if err != nil {
		return nil, err
	}

	_bar.Incr()

	err = ioutil.WriteFile(path.Join(args.TargetDir, "config.json"), info.ConfigData, 0777)
	if err != nil {
		return nil, err
	}

	_bar.Incr()

	manifest, err := getManifest(info.ConfigData, args.Platform)
	if err != nil {
		return nil, err
	}

	_bar.Incr()
	_barTotal.Incr()

	return manifest, nil
}

func getUpdateInfo(args cord.Args, usePatch *bool, gameVer string, updateOnly bool) (*models.UpdateInfo, error) {

	_curTitle = "Getting update info"

	var info *models.UpdateInfo
	var err error

	api := cordapi.NewCordAPI(args.Url)
	err = api.Login(args.Login, args.Password)
	if err != nil {
		return nil, err
	}

	if *usePatch {

		info, err = api.GetUpdatePatch(args.GameID, args.BranchName, args.Locale, args.Platform, gameVer)
		if err != nil {

			*usePatch = false // try to download whole build

		} else {

			_curTitle = "Checking"

			torrentFile := path.Join(args.TargetDir, "torrent.torrent")
			err := VerifyTorrentFile(torrentFile, args.TargetDir, _bar)
			if err != nil {

				*usePatch = false // try to download whole build
				if updateOnly {
					return nil, err
				}
			}
		}
	}

	if !*usePatch {

		_barTotal.Incr()

		info, err = api.GetUpdateInfo(args.GameID, args.BranchName, args.Locale, args.Platform, gameVer)
		if err != nil {
			return nil, err
		}
	}

	return info, nil
}

func doInstall(args cord.Args, manifest *models.ConfigManifest, steps int) error {

	_bar.Set(0)
	_bar.Total = 3
	_curTitle = "Installing game ..."

	if steps&redistNeed == redistNeed {
		err := downloadRedist(manifest)
		if err != nil {
			return err
		}
	}

	_bar.Incr()

	if steps&instralNeed == instralNeed {
		err := install(args.TargetDir, manifest)
		if err != nil {
			return err
		}
	}

	_bar.Incr()

	if steps&registyNeed == registyNeed {
		err := utils2.AddRegKeys(manifest)
		if err != nil {
			return err
		}
	}

	_bar.Incr()
	_barTotal.Incr()

	return nil
}

func isNeedInstall(manifestOld *models.ConfigManifest, manifestNew *models.ConfigManifest) int {

	steps := noNeed

	for i, rNew := range manifestNew.Redistributables {

		if rNew != manifestOld.Redistributables[i] {
			steps |= redistNeed
			break
		}
	}

	for i, scrNew := range manifestNew.InstallScripts {

		if !reflect.DeepEqual(scrNew, manifestOld.InstallScripts[i]) {
			steps |= instralNeed
			break
		}
	}

	for i, rkNew := range manifestNew.RegistryKeys {

		if !reflect.DeepEqual(rkNew, manifestOld.RegistryKeys[i]) {
			steps |= registyNeed
			break
		}
	}

	return steps
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
