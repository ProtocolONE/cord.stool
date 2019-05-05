package build

import (
	"fmt"

	"cord.stool/branch"
	"cord.stool/compressor/gzip"
	"cord.stool/context"
	"cord.stool/upload/cord"

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
					/*cli.BoolFlag{
						Name:        "patch, p",
						Usage:       "Uploads the difference between files using xdelta algorithm",
						Destination: &args.cordArgs.Patch,
					},
					cli.BoolFlag{
						Name:        "hash, h",
						Usage:       "Uploads changed files only",
						Destination: &args.cordArgs.Hash,
					},*/
					cli.BoolFlag{
						Name:        "wsync, w",
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
				},
				Action: func(c *cli.Context) error {
					return doPublish(ctx, c)
				},
			},
			cli.Command{
				Name:        "pull",
				Usage:       "Downloads build",
				Description: "Downloads build from Cord server",
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
			cli.Command{
				Name:        "live",
				Usage:       "Sets live build",
				Description: "Sets live build for specified branch",
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
						Name:        "build, bid",
						Usage:       "Build ID",
						Value:       "",
						Destination: &args.cordArgs.BuildID,
					},
				},
				Action: func(c *cli.Context) error {
					return doLive(ctx, c)
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

	err := cord.Upload(args.cordArgs)
	if err != nil {
		return err
	}

	return nil
}

func doPublish(ctx *context.StoolContext, c *cli.Context) error {

	return nil
}

func doUpdate(ctx *context.StoolContext, c *cli.Context) error {

	return nil
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

	err := branch.ListBuild(args.cordArgs.Url, args.cordArgs.Login, args.cordArgs.Password, args.cordArgs.GameID, args.cordArgs.BranchName)
	if err != nil {
		return err
	}

	return nil
}

func doLive(ctx *context.StoolContext, c *cli.Context) error {

	if args.cordArgs.Url == "" {
		return fmt.Errorf("-url flag is required")
	}

	if args.cordArgs.GameID == "" {
		return fmt.Errorf("Game ID is required")
	}

	if args.cordArgs.BranchName == "" {
		return fmt.Errorf("Branch name is required")
	}

	if args.cordArgs.BuildID == "" {
		return fmt.Errorf("Build ID is required")
	}

	err := branch.LiveBuild(args.cordArgs.Url, args.cordArgs.Login, args.cordArgs.Password, args.cordArgs.GameID, args.cordArgs.BranchName, args.cordArgs.BuildID)
	if err != nil {
		return err
	}

	return nil
}

func do(ctx *context.StoolContext, c *cli.Context) error {

	return fmt.Errorf("Specify one of following sub-commands: push, publish or update")
}
