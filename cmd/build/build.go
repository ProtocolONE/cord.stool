package build

import (
	"fmt"

	"cord.stool/branch"
	"cord.stool/compressor/gzip"
	"cord.stool/context"
	"cord.stool/upload/cord"
	"cord.stool/update"

	"github.com/urfave/cli"
)

var args = struct {
	cordArgs cord.Args
}{}

func Register(ctx *context.StoolContext) {

	cmd := cli.Command{
		Name:        "build",
		Usage:       "Manages builds",
		Description: "Manages builds",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "url",
				Usage:       "Cord server url",
				Value:       "",
				Destination: &args.cordArgs.Url,
			},
			cli.StringFlag{
				Name:        "login, l",
				Usage:       "Cord user login",
				Value:       "",
				Destination: &args.cordArgs.Login,
			},
			cli.StringFlag{
				Name:        "password, p",
				Usage:       "Cord user password",
				Value:       "",
				Destination: &args.cordArgs.Password,
			},
		},

		Subcommands: cli.Commands{
			cli.Command{
				Name:        "push",
				Usage:       "Uploads build",
				Description: "Uploads build to Cord server",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "source, s",
						Usage:       "Path to game",
						Value:       "",
						Destination: &args.cordArgs.SourceDir,
					},
					cli.StringFlag{
						Name:        "game-id, gid",
						Usage:       "Game ID",
						Value:       "",
						Destination: &args.cordArgs.GameID,
					},
					cli.StringFlag{
						Name:        "branch-name, bn",
						Usage:       "Branch name",
						Value:       "",
						Destination: &args.cordArgs.BranchName,
					},
					cli.BoolFlag{
						Name:        "force, f",
						Usage:       "Creates branch if it is not exits",
						Destination: &args.cordArgs.Force,
					},
					cli.StringFlag{
						Name:        "config, c",
						Usage:       "Path to build config file",
						Value:       "",
						Destination: &args.cordArgs.Config,
					},
					cli.BoolFlag{
						Name:        "sync",
						Usage:       "Uploads changed files only using Wharf protocol that enables incremental uploads",
						Destination: &args.cordArgs.Wharf,
					},
				},
				Action: func(c *cli.Context) error {
					return doPush(ctx, c)
				},
			},
			cli.Command{
				Name:        "publish",
				Usage:       "Publishes build",
				Description: "Publishes build to Cord server",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "game-id, gid",
						Usage:       "Game ID",
						Value:       "",
						Destination: &args.cordArgs.GameID,
					},
					cli.StringFlag{
						Name:        "branch-name, bn",
						Usage:       "Branch name",
						Value:       "",
						Destination: &args.cordArgs.BranchName,
					},
					/*cli.StringFlag{
						Name:        "build-id, bi",
						Usage:       "Build ID, optional",
						Value:       "",
						Destination: &args.cordArgs.BuildID,
					},*/
				},
				Action: func(c *cli.Context) error {
					return doPublish(ctx, c)
				},
			},
			cli.Command{
				Name:        "update",
				Usage:       "Downloads and install build",
				Description: "Downloads build from Cord server and install it",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "game-id, gid",
						Usage:       "Game ID",
						Value:       "",
						Destination: &args.cordArgs.GameID,
					},
					cli.StringFlag{
						Name:        "branch-name, bn",
						Usage:       "Branch name",
						Value:       "",
						Destination: &args.cordArgs.BranchName,
					},
					cli.StringFlag{
						Name:        "target, t",
						Usage:       "Path to install/update a game",
						Value:       "",
						Destination: &args.cordArgs.TargetDir,
					},
					cli.StringFlag{
						Name:        "locale, l",
						Usage:       "Locale [default: en-US]",
						Value:       "en-US",
						Destination: &args.cordArgs.Locale,
					},
					cli.StringFlag{
						Name:        "platform, p",
						Usage:       "Platform [default: win64]",
						Value:       "win64",
						Destination: &args.cordArgs.Platform,
					},
				},
				Action: func(c *cli.Context) error {
					return doUpdate(ctx, c)
				},
			},
			cli.Command{
				Name:        "list",
				Usage:       "Shows build list",
				Description: "Shows all builds specified branch",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "game-id, gid",
						Usage:       "Game ID",
						Value:       "",
						Destination: &args.cordArgs.GameID,
					},
					cli.StringFlag{
						Name:        "branch-name, bn",
						Usage:       "Branch name",
						Value:       "",
						Destination: &args.cordArgs.BranchName,
					},
				},
				Action: func(c *cli.Context) error {
					return doList(ctx, c)
				},
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)

	gzip.Init()
}

func doPush(ctx *context.StoolContext, c *cli.Context) error {

	if args.cordArgs.SourceDir == "" {
		return fmt.Errorf("-source flag is required")
	}

	if args.cordArgs.Url == "" {
		return fmt.Errorf("-url flag is required")
	}

	if args.cordArgs.GameID == "" {
		return fmt.Errorf("Game ID is required")
	}

	if args.cordArgs.BranchName == "" {
		return fmt.Errorf("Branch name is required")
	}

	if args.cordArgs.Config == "" {
		return fmt.Errorf("Config file is required")
	}

	return cord.Upload(args.cordArgs)
}

func doUpdate(ctx *context.StoolContext, c *cli.Context) error {

	if args.cordArgs.Url == "" {
		return fmt.Errorf("-url flag is required")
	}

	if args.cordArgs.GameID == "" {
		return fmt.Errorf("Game ID is required")
	}

	if args.cordArgs.BranchName == "" {
		return fmt.Errorf("Branch name is required")
	}

	if args.cordArgs.TargetDir == "" {
		return fmt.Errorf("-target flag is required")
	}

	return update.Update(args.cordArgs)
}

func doList(ctx *context.StoolContext, c *cli.Context) error {

	if args.cordArgs.Url == "" {
		return fmt.Errorf("-url flag is required")
	}

	if args.cordArgs.GameID == "" {
		return fmt.Errorf("Game ID is required")
	}

	if args.cordArgs.BranchName == "" {
		return fmt.Errorf("Branch name is required")
	}

	return branch.ListBuild(args.cordArgs.Url, args.cordArgs.Login, args.cordArgs.Password, args.cordArgs.GameID, args.cordArgs.BranchName)
}

func doPublish(ctx *context.StoolContext, c *cli.Context) error {

	if args.cordArgs.Url == "" {
		return fmt.Errorf("-url flag is required")
	}

	if args.cordArgs.GameID == "" {
		return fmt.Errorf("Game ID is required")
	}

	if args.cordArgs.BranchName == "" {
		return fmt.Errorf("Branch name is required")
	}

	return branch.PublishBuild(args.cordArgs.Url, args.cordArgs.Login, args.cordArgs.Password, args.cordArgs.GameID, args.cordArgs.BranchName, args.cordArgs.BuildID)
}

func do(ctx *context.StoolContext, c *cli.Context) error {

	return fmt.Errorf("Specify one of following sub-commands: push, publish, list or update")
}
