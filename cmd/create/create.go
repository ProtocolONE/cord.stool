package create

import (
	"cord.stool/context"
	"cord.stool/updater"
	"github.com/urfave/cli"
)

var args = struct {
	SourceDir string
	TargetDir string
	Archive   bool
}{
	Archive: false,
}

func Register(ctx *context.StoolContext) {
	cmd := cli.Command{
		Name:        "create",
		ShortName:   "c",
		Usage:       "Create update",
		Description: "Create update for application",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "source, s",
				Usage:       "Path to game",
				Value:       "",
				Destination: &args.SourceDir,
			},
			cli.StringFlag{
				Name:        "output, o",
				Usage:       "Path to game",
				Value:       "",
				Destination: &args.TargetDir,
			},
			cli.BoolFlag{
				Name:        "archive, a",
				Usage:       "Archie with zip each file",
				Destination: &args.Archive,
			},
		},
		Action: func(c *cli.Context) error {
			e := do(ctx, c)

			return e
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func do(ctx *context.StoolContext, c *cli.Context) error {
	_, e := updater.PrepairDistr(args.SourceDir, args.TargetDir, args.Archive)

	return e
}
