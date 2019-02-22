package torrent

import (
	"os"
	"time"

	"fmt"
	"io"
	"path/filepath"
	"strings"

	"cord.stool/context"
	"cord.stool/cordapi"
	"cord.stool/service/models"
	"cord.stool/utils"

	"github.com/anacrolix/missinggo/slices"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gosuri/uiprogress"
	"github.com/gosuri/uiprogress/util/strutil"
	"github.com/urfave/cli"
)

var args = struct {
	SourceDir    string
	TargetFile   string
	WebSeeds     cli.StringSlice
	AnnounceList cli.StringSlice
	PieceLength  int64
	Url          string
	Login        string
	Password     string
}{}

var progrssBar *uiprogress.Bar
var curProgressTitle string

func Register(ctx *context.StoolContext) {

	cmd := cli.Command{
		Name:        "torrent",
		ShortName:   "t",
		Usage:       "Create torrent",
		Description: "Create torrent file and adding it to Torrent Tracker if needed it",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "source, s",
				Usage:       "Path to game",
				Value:       "",
				Destination: &args.SourceDir,
			},
			cli.StringFlag{
				Name:        "file, f",
				Usage:       "Path to torrent file",
				Value:       "",
				Destination: &args.TargetFile,
			},
			cli.StringSliceFlag{
				Name:  "web-seeds, ws",
				Value: &args.WebSeeds,
				Usage: "Slice of torrent web seeds",
			},
			cli.StringSliceFlag{
				Name:  "announce-list, al",
				Value: &args.AnnounceList,
				Usage: "Slice of announce server url",
			},
			cli.Int64Flag{
				Name:        "piece-length, pl",
				Usage:       "Torrent piece length",
				Value:       512,
				Destination: &args.PieceLength,
			},
			cli.StringFlag{
				Name:        "cord-url",
				Usage:       "Cord server url",
				Value:       "",
				Destination: &args.Url,
			},
			cli.StringFlag{
				Name:        "cord-login",
				Usage:       "Cord user login",
				Value:       "",
				Destination: &args.Login,
			},
			cli.StringFlag{
				Name:        "cord-password",
				Usage:       "Cord user password",
				Value:       "",
				Destination: &args.Password,
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

var (
	builtinAnnounceList = [][]string{
		//{"udp://tracker.openbittorrent.com:80"},
		// {"udp://tracker.publicbt.com:80"},
		// {"udp://tracker.istole.it:6969"},
	}
)

// This is a helper that sets Files and Pieces from a root path and its
// children.
func buildFromFilePathEx(root string, ignoreFiles map[string]bool) (info metainfo.Info, err error) {

	info = metainfo.Info{
		PieceLength: args.PieceLength * 1024,
		Name:        "live",
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

func createTorrent(rootDir string, targetFile string, announceList []string, urlList []string) (err error) {

	fmt.Println("Creating torrent file ...")

	fc, err := utils.FileCount(rootDir)
	if err != nil {
		return err
	}

	uiprogress.Start()
	progrssBar = uiprogress.AddBar(fc + 4).AppendCompleted().PrependElapsed()

	var title *string
	title = &curProgressTitle
	curProgressTitle = "Getting metainfo ..."

	progrssBar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(*title, 35)
	})

	mi := metainfo.MetaInfo{
		AnnounceList: builtinAnnounceList,
		CreatedBy:    "stool",
		CreationDate: time.Now().Unix(),
	}

	for _, a := range announceList {
		mi.AnnounceList = append(mi.AnnounceList, []string{a})
	}

	for _, u := range urlList {
		mi.UrlList = append(mi.UrlList, u)
	}

	ignoreFiles := map[string]bool{
		"update.crc.zip": true,
		"update.crc":     true,
	}

	info, err := buildFromFilePathEx(rootDir, ignoreFiles)
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
	uiprogress.Stop()

	fmt.Println("Creating is completed.")

	return
}

func do(ctx *context.StoolContext, c *cli.Context) error {

	if args.TargetFile == "" {
		return fmt.Errorf("Path to torrent file is required")
	}

	if args.SourceDir == "" && args.Url == "" {
		return fmt.Errorf("Specify one of following flags: source or cord-url")
	}

	if args.SourceDir != "" {

		err := createTorrent(
			args.SourceDir,
			args.TargetFile,
			args.AnnounceList,
			args.WebSeeds)

		if err != nil {
			return err
		}
	}

	if args.Url != "" {

		err := addTorrent(args.Url, args.Login, args.Password, args.TargetFile)
		if err != nil {
			return err
		}

	}

	return nil
}

func addTorrent(url, login, password, torrent string) error {

	fmt.Println("Adding torrent to Cord Server ...")

	uiprogress.Start()
	progrssBar := uiprogress.AddBar(3).AppendCompleted().PrependElapsed()

	var title *string
	title = &curProgressTitle
	curProgressTitle = "Getting metainfo from torent file..."

	progrssBar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(*title, 35)
	})

	mi, err := metainfo.LoadFromFile(torrent)
	if err != nil {
		return err
	}

	hash := metainfo.HashBytes(mi.InfoBytes)
	infoHash := hash.String()

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
