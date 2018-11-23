package torrent

import (
	"os"
	"time"

	"cord.stool/context"
	"github.com/urfave/cli"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

var args = struct {
	SourceDir    string
	TargetFile   string
	WebSeeds     cli.StringSlice
	AnnounceList cli.StringSlice
}{}

func Register(ctx *context.StoolContext) {
	cmd := cli.Command{
		Name:        "torrent",
		Usage:       "Create torrent",
		Description: "Create torrent file",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "sourceDir, sd",
				Usage:       "Path to game(short form sd)",
				Value:       "",
				Destination: &args.SourceDir,
			},

			cli.StringFlag{
				Name:        "targetFile, tf",
				Usage:       "Path for new torrent file",
				Value:       "",
				Destination: &args.TargetFile,
			},
			cli.StringSliceFlag{
				Name:   "web-seeds, ws",
				Value: &args.WebSeeds,
				Usage: "Slice of torrent web seeds",
			},
			cli.StringSliceFlag{
				Name:  "announce-list, a",
				Value: &args.AnnounceList,
				Usage: "Slice of announce server url",
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
		{"udp://tracker.openbittorrent.com:80"},
		// {"udp://tracker.publicbt.com:80"},
		// {"udp://tracker.istole.it:6969"},
	}
)

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

	info := metainfo.Info{
		// UNDONE Should we add here auto detect best piece length by maximum torrent size?
		PieceLength: 256 * 1024,
	}

	err = info.BuildFromFilePath(rootDir)
	if err != nil {
		return
	}

	// INFO It's imortant to be `live` for current GameDownloader
	info.Name = "live"

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
