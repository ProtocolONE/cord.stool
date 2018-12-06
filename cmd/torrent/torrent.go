package torrent

import (
	"os"
	"time"

	"fmt"
	"io"
	"path/filepath"
	"strings"

	"cord.stool/context"
	"github.com/urfave/cli"

	"github.com/anacrolix/missinggo/slices"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

var args = struct {
	SourceDir    string
	TargetFile   string
	WebSeeds     cli.StringSlice
	AnnounceList cli.StringSlice
	PieceLength  int64
}{}

func Register(ctx *context.StoolContext) {
	cmd := cli.Command{
		Name:        "torrent",
		Usage:       "Create torrent",
		Description: "Create torrent file",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "source, s",
				Usage:       "Path to game",
				Value:       "",
				Destination: &args.SourceDir,
			},
			cli.StringFlag{
				Name:        "target, t",
				Usage:       "Path for new torrent file",
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
func BuildFromFilePathEx(root string, ignoreFiles map[string]bool) (info metainfo.Info, err error) {
	info = metainfo.Info{
		PieceLength: args.PieceLength * 1024,
		Name:        "live",
		Files:       nil,
	}

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
	slices.Sort(info.Files, func(l, r metainfo.FileInfo) bool {
		return strings.Join(l.Path, "/") < strings.Join(r.Path, "/")
	})
	err = info.GeneratePieces(func(fi metainfo.FileInfo) (io.ReadCloser, error) {
		return os.Open(filepath.Join(root, strings.Join(fi.Path, string(filepath.Separator))))
	})
	if err != nil {
		err = fmt.Errorf("error generating pieces: %s", err)
	}
	return
}

func CreateTorrent(rootDir string, targetFile string, announceList []string, urlList []string) (err error) {
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

	info, err := BuildFromFilePathEx(rootDir, ignoreFiles)
	if err != nil {
		return
	}

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

	return
}

func do(ctx *context.StoolContext, c *cli.Context) error {
	return CreateTorrent(
		args.SourceDir,
		args.TargetFile,
		args.AnnounceList,
		args.WebSeeds)
}
