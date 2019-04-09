package branch

import (
	"fmt"

	"cord.stool/branch"
	"cord.stool/context"
	"github.com/urfave/cli"
)

var args = struct {
	nameOrID  string
	gameID    string
	sNameOrID string
	tNameOrID string
	url       string
	login     string
	password  string
}{}

func Register(ctx *context.StoolContext) {

	cmd := cli.Command{
		Name:        "branch",
		ShortName:   "b",
		Usage:       "Manage branches",
		Description: "Manage branches",

		Subcommands: cli.Commands{
			cli.Command{
				Name:        "create",
				ShortName:   "c",
				Usage:       "Creates branch",
				Description: "Creates branch",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "name, n",
						Usage:       "Branch name or branch ID",
						Value:       "",
						Destination: &args.nameOrID,
					},
					cli.StringFlag{
						Name:        "game-id, gid",
						Usage:       "Game ID",
						Value:       "",
						Destination: &args.gameID,
					},
				},
				Action: func(c *cli.Context) error {
					return doCreate(ctx, c)
				},
			},
			cli.Command{
				Name:        "delete",
				ShortName:   "d",
				Usage:       "Deletes branch",
				Description: "Deletes branch",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "name, n",
						Usage:       "Branch name or branch ID",
						Value:       "",
						Destination: &args.nameOrID,
					},
					cli.StringFlag{
						Name:        "game-id, gid",
						Usage:       "Game ID",
						Value:       "",
						Destination: &args.gameID,
					},
				},
				Action: func(c *cli.Context) error {
					return doDelete(ctx, c)
				},
			},
			cli.Command{
				Name:        "list",
				ShortName:   "l",
				Usage:       "Shows branch",
				Description: "Shows branch",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "game-id, gid",
						Usage:       "Game ID",
						Value:       "",
						Destination: &args.gameID,
					},
				},
				Action: func(c *cli.Context) error {
					return doList(ctx, c)
				},
			},
			cli.Command{
				Name:        "shallow",
				ShortName:   "s",
				Usage:       "Shallows branch",
				Description: "Shallows branch",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "sname, sn",
						Usage:       "Source branch name or branch ID",
						Value:       "",
						Destination: &args.sNameOrID,
					},
					cli.StringFlag{
						Name:        "tname, tn",
						Usage:       "Target branch name or branch ID",
						Value:       "",
						Destination: &args.tNameOrID,
					},
				},
				Action: func(c *cli.Context) error {
					return doShallow(ctx, c)
				},
			},
		},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "url",
				Usage:       "Cord server url",
				Value:       "",
				Destination: &args.url,
			},
			cli.StringFlag{
				Name:        "login",
				Usage:       "Cord user login",
				Value:       "",
				Destination: &args.login,
			},
			cli.StringFlag{
				Name:        "password",
				Usage:       "Cord user password",
				Value:       "",
				Destination: &args.password,
			},
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func doCreate(ctx *context.StoolContext, c *cli.Context) error {

	if args.url == "" {
		return fmt.Errorf("Cord server url is required")
	}

	return branch.CreateBranch(args.url, args.login, args.password, args.gameID, args.nameOrID)
}

func doDelete(ctx *context.StoolContext, c *cli.Context) error {

	if args.url == "" {
		return fmt.Errorf("Cord server url is required")
	}

	return branch.DeleteBranch(args.url, args.login, args.password, args.gameID, args.nameOrID)
}

func doList(ctx *context.StoolContext, c *cli.Context) error {

	if args.url == "" {
		return fmt.Errorf("Cord server url is required")
	}

	return branch.ListBranch(args.url, args.login, args.password, args.gameID)
}

func doShallow(ctx *context.StoolContext, c *cli.Context) error {

	if args.url == "" {
		return fmt.Errorf("Cord server url is required")
	}

	return branch.ShallowBranch(args.url, args.login, args.password, args.sNameOrID, args.tNameOrID)
}
