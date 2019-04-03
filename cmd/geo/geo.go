package geo

import (
	"fmt"

	"cord.stool/context"
	"cord.stool/geo"

	"github.com/urfave/cli"
)

var args = struct {
	host     string
	port     string
	password string
	db       int
	blocks   string
}{}

func Register(ctx *context.StoolContext) {

	cmd := cli.Command{
		Name:        "geo",
		ShortName:   "g",
		Usage:       "Import geo data",
		Description: "Import geo database to redis",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "host",
				Usage:       "Redis host",
				Value:       "",
				Destination: &args.host,
			},
			cli.StringFlag{
				Name:        "port",
				Usage:       "Redis port",
				Value:       "",
				Destination: &args.port,
			},
			cli.StringFlag{
				Name:        "password, p",
				Usage:       "Redis password",
				Value:       "",
				Destination: &args.password,
			},
			cli.IntFlag{
				Name:        "db",
				Usage:       "DB index",
				Value:       0,
				Destination: &args.db,
			},
			cli.StringFlag{
				Name:        "import-blocks, ib",
				Usage:       "Imports the location blocks file, import-blocks <file>.",
				Value:       "",
				Destination: &args.blocks,
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func do(ctx *context.StoolContext, c *cli.Context) error {

	if args.host == "" || args.port == "" {
		return fmt.Errorf("Specify redis host and port")
	}

	if args.blocks == "" {
		return fmt.Errorf("Specify one of following flags: import-blocks")
	}

	client := geo.NewGeoClient(args.host, args.port, args.password, args.db)

	if args.blocks != "" {
		err := client.ImportBlocks(args.blocks)
		if err != nil {
			return err
		}
	}

	return nil
}
