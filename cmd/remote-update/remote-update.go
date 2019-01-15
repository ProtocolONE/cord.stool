package remote_update

import (
	"fmt"

	"cord.stool/context"
	"cord.stool/upload/cord"

	"github.com/urfave/cli"
)

var args = struct {
	Url    string
	Login    string
	Password    string
	SourceDir string
	OutputDir string
}{}

func Register(ctx *context.StoolContext) {
	cmd := cli.Command{
		Name:        	"remote-update",
		ShortName:		"p",
		Usage:			"Upload update to cord server",
		Description:	"Upload update app bundle to cord server",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "url, u",
				Usage:       "Server url",
				Value:       "",
				Destination: &args.Url,
			},
			cli.StringFlag{
				Name:        "login, l",
				Usage:       "User login",
				Value:       "",
				Destination: &args.Login,
			},
			cli.StringFlag{
				Name:        "password, p",
				Usage:       "User password",
				Value:       "",
				Destination: &args.Password,
			},
			cli.StringFlag{
				Name:        "source, s",
				Usage:       "Path to game",
				Value:       "",
				Destination: &args.SourceDir,
			},
			cli.StringFlag{
				Name:        "output, o",
				Usage:       "Server storage name to upload",
				Value:       "",
				Destination: &args.OutputDir,
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func do(ctx *context.StoolContext, c *cli.Context) error {
	
	if args.Url == "" {
		return fmt.Errorf("Server Url is required")
	}

	if args.Login == "" {
		return fmt.Errorf("User login is required")
	}

	if args.Password == "" {
		return fmt.Errorf("User password is required")
	}

	if args.SourceDir == "" {
		return fmt.Errorf("Path to game is required")
	}

	if args.OutputDir == "" {
		return fmt.Errorf("Server storage name is required")
	}

	return cord.Upload(args.Url, args.Login, args.Password, args.SourceDir, args.OutputDir)
}
