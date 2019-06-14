package torrent

import (
	"cord.stool/context"
	"cord.stool/upload/cord"
	"fmt"
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

func do(ctx *context.StoolContext, c *cli.Context) error {

	if args.TargetFile == "" {
		return fmt.Errorf("Path to torrent file is required")
	}

	if args.SourceDir == "" && args.Url == "" {
		return fmt.Errorf("Specify one of following flags: source or cord-url")
	}

	if args.SourceDir != "" {

		ignoreFiles := map[string]bool{}

		err := cord.CreateTorrent(
			args.SourceDir,
			args.TargetFile,
			args.AnnounceList,
			args.WebSeeds,
			ignoreFiles,
			args.PieceLength,
			true,
		)

		if err != nil {
			return err
		}
	}

	if args.Url != "" {

		err := cord.AddTorrent(args.Url, args.Login, args.Password, args.TargetFile)
		if err != nil {
			return err
		}
	}

	return nil
}
