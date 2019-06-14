package cord

import (
	"os"
	"time"

	"fmt"
	"io"
	"path/filepath"
	"strings"

	"cord.stool/cordapi"
	"cord.stool/service/models"
	"cord.stool/utils"

	"github.com/anacrolix/missinggo/slices"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
)

var progrssBar *uiprogress.Bar
var curProgressTitle string

// This is a helper that sets Files and Pieces from a root path and its
// children.
func buildFromFilePathEx(root string, ignoreFiles map[string]bool, pieceLength int64) (info metainfo.Info, err error) {

	info = metainfo.Info{
		PieceLength: pieceLength * 1024,
		Name:        "content",
		Files:       nil,
	}

	progrssBar.Incr()
	curProgressTitle = "Getting files info ..."

	err = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {

		if err != nil {
			return err
		}
		if fi.IsDir() {
			// Directories are implicit in torrent files.
			return nil
		} else if path == root {
			// The root is a file.
			info.Length = fi.Size()
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("error getting relative path: %s", err)
		}

		if _, found := ignoreFiles[relPath]; found {
			return nil
		}

		info.Files = append(info.Files, metainfo.FileInfo{
			Path:   strings.Split(relPath, string(filepath.Separator)),
			Length: fi.Size(),
		})
		return nil
	})

	if err != nil {
		return
	}

	progrssBar.Incr()
	curProgressTitle = "Generating pieces ..."

	slices.Sort(info.Files, func(l, r metainfo.FileInfo) bool {
		return strings.Join(l.Path, "/") < strings.Join(r.Path, "/")
	})

	err = info.GeneratePieces(func(fi metainfo.FileInfo) (io.ReadCloser, error) {
		progrssBar.Incr()
		return os.Open(filepath.Join(root, strings.Join(fi.Path, string(filepath.Separator))))
	})

	if err != nil {
		err = fmt.Errorf("error generating pieces: %s", err)
	}

	return
}

func CreateTorrent(rootDir string, targetFile string, announceList []string, urlList []string, ignoreFiles map[string]bool, pieceLength int64, silent bool) (err error) {

	if !silent {
		fmt.Println("Creating torrent file ...")
	}

	fc, err := utils.FileCount(rootDir)
	if err != nil {
		return err
	}

	if !silent {
		uiprogress.Start()
	}

	progrssBar = uiprogress.AddBar(fc + 4).AppendCompleted().PrependElapsed()

	var title *string
	title = &curProgressTitle
	curProgressTitle = "Getting metainfo ..."

	progrssBar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(*title, 35)
	})

	mi := metainfo.MetaInfo{
		CreatedBy:    "cord.stool",
		CreationDate: time.Now().Unix(),
	}

	for _, a := range announceList {
		mi.AnnounceList = append(mi.AnnounceList, []string{a})
	}

	for _, u := range urlList {
		mi.UrlList = append(mi.UrlList, u)
	}

	info, err := buildFromFilePathEx(rootDir, ignoreFiles, pieceLength)
	if err != nil {
		return
	}

	progrssBar.Incr()
	curProgressTitle = "Creating torrent file ..."

	mi.InfoBytes, err = bencode.Marshal(info)
	if err != nil {
		return
	}

	f, err := os.Create(targetFile)
	if err != nil {
		return
	}

	defer f.Close()
	err = mi.Write(f)

	progrssBar.Incr()
	curProgressTitle = "Finished"
	title = &curProgressTitle

	if !silent {
		uiprogress.Stop()
		fmt.Println("Creating is completed.")
	}

	return
}

func GetInfoHash(torrent string) (string, error) {

	mi, err := metainfo.LoadFromFile(torrent)
	if err != nil {
		return "", err
	}

	hash := metainfo.HashBytes(mi.InfoBytes)
	return hash.String(), nil
}

func AddTorrent(url, login, password, torrent string) error {

	fmt.Println("Adding torrent to Cord Server ...")

	uiprogress.Start()
	progrssBar := uiprogress.AddBar(3).AppendCompleted().PrependElapsed()

	var title *string
	title = &curProgressTitle
	curProgressTitle = "Getting metainfo from torent file..."

	progrssBar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(*title, 35)
	})

	infoHash, err := GetInfoHash(torrent)
	if err != nil {
		return err
	}

	progrssBar.Incr()
	curProgressTitle = "Login to Torrent Tracker"

	api := cordapi.NewCordAPI(url)
	err = api.Login(login, password)
	if err != nil {
		return err
	}

	progrssBar.Incr()
	curProgressTitle = "Adding torrent to Torrent Tracker"

	err = api.AddTorrent(&models.TorrentCmd{infoHash})
	if err != nil {
		return err
	}

	progrssBar.Incr()
	curProgressTitle = "Finished"
	title = &curProgressTitle
	uiprogress.Stop()

	fmt.Println("Torrent is added to Torrent Tracker.")

	return nil
}
