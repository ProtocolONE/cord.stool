package branch

import (
	"fmt"

	"cord.stool/branch"
	"cord.stool/context"
	"github.com/urfave/cli"
)

var args = struct {
	id       string
	name     string
	gameID   string
	sID      string
	tID      string
	sName    string
	tName    string
	url      string
	login    string
	password string
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
						Usage:       "Branch name",
						Value:       "",
						Destination: &args.name,
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
						Name:        "id",
						Usage:       "Branch ID",
						Value:       "",
						Destination: &args.id,
					},
					cli.StringFlag{
						Name:        "name, n",
						Usage:       "Branch name. Should be specified with game id",
						Value:       "",
						Destination: &args.name,
					},
					cli.StringFlag{
						Name:        "game-id, gid",
						Usage:       "Game ID. Should be specified with branch name",
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
				Usage:       "Shows branch list",
				Description: "Shows branch list",
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
						Name:        "source-id, sid",
						Usage:       "Source branch ID",
						Value:       "",
						Destination: &args.sID,
					},
					cli.StringFlag{
						Name:        "target-id, tid",
						Usage:       "Target branch ID",
						Value:       "",
						Destination: &args.tID,
					},
					cli.StringFlag{
						Name:        "source-name, sn",
						Usage:       "Source branch name. Should be specified with game id",
						Value:       "",
						Destination: &args.sName,
					},
					cli.StringFlag{
						Name:        "target-name, tn",
						Usage:       "Target branch name. Should be specified with game id",
						Value:       "",
						Destination: &args.tName,
					},
					cli.StringFlag{
						Name:        "game-id, gid",
						Usage:       "Game ID. Should be specified with branch names",
						Value:       "",
						Destination: &args.gameID,
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

	return branch.CreateBranch(args.url, args.login, args.password, args.name, args.gameID)
}

func doDelete(ctx *context.StoolContext, c *cli.Context) error {

	if args.url == "" {
		return fmt.Errorf("Cord server url is required")
	}

	return branch.DeleteBranch(args.url, args.login, args.password, args.id, args.name, args.gameID)
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

	return branch.ShallowBranch(args.url, args.login, args.password, args.sID, args.sName, args.tID, args.tName, args.gameID)
}
