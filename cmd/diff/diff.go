package diff

import (
	"fmt"

	"cord.stool/context"
	"cord.stool/updater"

	"github.com/urfave/cli"
)

var args = struct {
	SourceOldDir  string
	SourceNewDir  string
	OutputDiffDir string
}{}

func Register(ctx *context.StoolContext) {
	cmd := cli.Command{
		Name:        "diff",
		ShortName:   "d",
		Usage:       "Make patch",
		Description: "Generate the difference between files",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "old, o",
				Usage:       "Path to old files",
				Value:       "",
				Destination: &args.SourceOldDir,
			},
			cli.StringFlag{
				Name:        "new, n",
				Usage:       "Path to new files",
				Value:       "",
				Destination: &args.SourceNewDir,
			},
			cli.StringFlag{
				Name:        "patch, p",
				Usage:       "Path to patch files",
				Value:       "",
				Destination: &args.OutputDiffDir,
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func do(ctx *context.StoolContext, c *cli.Context) error {

	if args.SourceOldDir == "" {
		return fmt.Errorf("SourceOldDir value required")
	} else if args.SourceNewDir == "" {
		return fmt.Errorf("SourceNewDir value required")
	} else if args.OutputDiffDir == "" {
		return fmt.Errorf("OutputDir value required")
	}

	return updater.CreateBinDiff(args.SourceOldDir, args.SourceNewDir, args.OutputDiffDir)
}
