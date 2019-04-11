package games

import (
	"fmt"

	"cord.stool/games"
	"cord.stool/context"
	"github.com/urfave/cli"
)

var args = struct {
	/*nameOrID  string
	gameID    string
	sNameOrID string
	tNameOrID string*/
	url       string
	login     string
	password  string
}{}

func Register(ctx *context.StoolContext) {

	cmd := cli.Command{
		Name:        "games",
		ShortName:   "g",
		Usage:       "Manage games",
		Description: "Manage games",

		Subcommands: cli.Commands{
			cli.Command{
				Name:        "list",
				ShortName:   "l",
				Usage:       "Shows game list",
				Description: "Shows game  list",
				/*Flags: []cli.Flag{
					cli.StringFlag{
						Name:        "game-id, gid",
						Usage:       "Game ID",
						Value:       "",
						Destination: &args.gameID,
					},
				},*/
				Action: func(c *cli.Context) error {
					return doList(ctx, c)
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

func doList(ctx *context.StoolContext, c *cli.Context) error {

	if args.url == "" {
		return fmt.Errorf("Cord server url is required")
	}

	return games.ListGame(args.url, args.login, args.password)
}
