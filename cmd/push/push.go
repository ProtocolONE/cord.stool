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
				Name:        "sourceDir, sd",
				Usage:       "Path to game(short form sd)",
				Value:       "",
				Destination: &args.SourceDir,
			},
			cli.StringFlag{
				Name:        "ftp",
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
