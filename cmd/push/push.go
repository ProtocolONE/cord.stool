package push

import (
	"fmt"

	"cord.stool/context"
	"cord.stool/upload/ftp"

	"github.com/urfave/cli"
)

var args = struct {
	FtpUrl    string
	SourceDir string
}{}

func Register(ctx *context.StoolContext) {
	cmd := cli.Command{
		Name:        "push",
		Usage:       "Upload update",
		Description: "Upload update app bundle to one of servers",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "source, s",
				Usage:       "Path to game",
				Value:       "",
				Destination: &args.SourceDir,
			},
			cli.StringFlag{
				Name:        "ftp",
				Usage:       "Full ftp url path. Example ftp://user@password:host:port/upload/directory",
				Value:       "",
				Destination: &args.FtpUrl,
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func do(ctx *context.StoolContext, c *cli.Context) error {
	if args.FtpUrl == "" {
		return fmt.Errorf("ftp url required")
	}

	return ftp.UploadToFTP(args.FtpUrl, args.SourceDir)
}
