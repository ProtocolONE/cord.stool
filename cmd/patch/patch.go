package patch

import (
	"fmt"

	"cord.stool/context"
	"cord.stool/update"

	"github.com/urfave/cli"
)

var args = struct {
	SourceOldDir string
	SourceNewDir string
	PatchFile    string
}{}

func Register(ctx *context.StoolContext) {
	cmd := cli.Command{
		Name:        "patch",
		ShortName:   "p",
		Usage:       "Make patch",
		Description: "Generate the patch file",

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
				Name:        "patch, pf",
				Usage:       "Path to patch file",
				Value:       "",
				Destination: &args.PatchFile,
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
		return fmt.Errorf("Source old dir value required")
	} else if args.SourceNewDir == "" {
		return fmt.Errorf("Source new dir value required")
	} else if args.PatchFile == "" {
		return fmt.Errorf("Output patch file value required")
	}

	return update.CreatePatch(args.SourceOldDir, args.SourceNewDir, args.PatchFile)
}
