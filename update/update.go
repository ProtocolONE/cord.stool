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
	humanize "github.com/dustin/go-humanize"
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

	usePatch := false
	gameVer := getGameVersion(args.TargetDir, args.Platform)
	if gameVer != "" {
		usePatch = true
	}

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

	_curTitle = "Getting update info"
	api := cordapi.NewCordAPI(args.Url)
	err := api.Login(args.Login, args.Password)
	if err != nil {
		return err
	}

	_barTotal.Incr()
	_bar.Incr()

	var info *models.UpdateInfo
	contentPath := path.Join(args.TargetDir, "content")

	if usePatch {

		info, err = api.GetUpdatePatch(args.GameID, args.BranchName, args.Locale, args.Platform, gameVer)
		if err != nil {
			usePatch = false // try to download whole build
		}
	}

	if !usePatch {

		_barTotal.Incr()

		info, err = api.GetUpdateInfo(args.GameID, args.BranchName, args.Locale, args.Platform, gameVer)
		if err != nil {
			return err
		}
	}

	_barTotal.Incr()
	_bar.Incr()

	torrentFile := path.Join(args.TargetDir, "torrent.torrent")

	stats := NewDownLoadStatistics("update")

	if gameVer == info.Version {

		_curTitle = "Checking"

		err = VerifyTorrentFile(torrentFile, args.TargetDir, _bar)
		if err != nil {

			_curTitle = "Downloading"
			err = StartDownLoadFile(torrentFile, args.TargetDir, _bar, stats)
			if err != nil {
				return err
			}
		}

	} else {

		_curTitle = "Downloading"
		err = StartDownLoad(info.TorrentData, args.TargetDir, _bar, stats)
		if err != nil {
			return err
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
			return err
		}

		_barTotal.Incr()
	}

	_bar.Set(0)
	_bar.Total = 3
	_curTitle = "Preparing to install ..."

	err = ioutil.WriteFile(torrentFile, info.TorrentData, 0777)
	if err != nil {
		return err
	}

	_bar.Incr()

	err = ioutil.WriteFile(path.Join(args.TargetDir, "config.json"), info.ConfigData, 0777)
	if err != nil {
		return err
	}

	_bar.Incr()

	cfg, err := utils.ReadConfigData(info.ConfigData, nil)
	if err != nil {
		return err
	}

	_bar.Incr()
	_barTotal.Incr()

	var manifest *models.ConfigManifest
	manifest = nil

	for _, m := range cfg.Application.Manifests {

		if m.Platform == args.Platform {
			manifest = &m
			break
		}
	}

	_bar.Set(0)
	_bar.Total = 3
	_curTitle = "Installing game ..."

	downloadRedist(manifest)
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

	_curTitle = "Finished"
	uiprogress.Stop()

	fmt.Println("Update completed.")
	//fmt.Printf("DownLoad statistics, time^ %s, max speed: %s/s", stats.DownloadTime.Format("2006-01-02 15:04:05 -0700"), humanize.IBytes(stats.MaxDownloadSpeed))

	return nil
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

		url := ""

		switch r {
		case models.Directx_june_2010:
			url = "https://download.microsoft.com/download/8/4/A/84A35BF1-DAFE-4AE8-82AF-AD2AE20B6B14/directx_Jun2010_redist.exe"
		case models.Vcredist_2005_x86:
			url = "https://download.microsoft.com/download/d/3/4/d342efa6-3266-4157-a2ec-5174867be706/vcredist_x86.exe"
		case models.Vcredist_2008_sp1_x86:
			url = "https://download.microsoft.com/download/d/d/9/dd9a82d0-52ef-40db-8dab-795376989c03/vcredist_x86.exe"
		case models.Vcredist_2010_x64:
			url = "https://download.microsoft.com/download/3/2/2/3224B87F-CFA0-4E70-BDA3-3DE650EFEBA5/vcredist_x64.exe"
		case models.Vcredist_2010_x86:
			url = "https://download.microsoft.com/download/5/B/C/5BC5DBB3-652D-4DCE-B14A-475AB85EEF6E/vcredist_x86.exe"
		case models.Vcredist_2012_update_4_x64:
			url = ""
		case models.Vcredist_2012_update_4_x86:
			url = ""
		case models.Vcredist_2013_x64:
			url = "https://aka.ms/highdpimfc2013x64enu"
		case models.Vcredist_2013_x86:
			url = "https://aka.ms/highdpimfc2013x86enu"
		case models.Vcredist_2015_x64:
			url = "https://aka.ms/vs/16/release/vc_redist.x64.exe"
		case models.Vcredist_2015_x86:
			url = "https://aka.ms/vs/16/release/vc_redist.x86.exe"
		case models.Vcredist_2017_x64:
			url = "https://aka.ms/vs/16/release/vc_redist.x64.exe"
		case models.Vcredist_2017_x86:
			url = "https://aka.ms/vs/16/release/vc_redist.x86.exe"
		case models.Xnafx_40:
			url = "https://download.microsoft.com/download/A/C/2/AC2C903B-E6E8-42C2-9FD7-BEBAC362A930/xnafx40_redist.msi"
		}

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
