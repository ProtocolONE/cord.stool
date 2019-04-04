package geo

import (
	"fmt"

	"cord.stool/context"
	"cord.stool/geo"

	"github.com/urfave/cli"
)

var args = struct {
	host        string
	port        int
	password    string
	db          int
	keyIPv4     string
	keyIPv6     string
	keyIPv6Info string
	blocks      string
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
				Value:       "localhost",
				Destination: &args.host,
			},
			cli.IntFlag{
				Name:        "port",
				Usage:       "Redis port",
				Value:       6379,
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
				Name:        "ip4",
				Usage:       "IPv4 redis key for IPv4 location",
				Value:       "",
				Destination: &args.keyIPv4,
			},
			cli.StringFlag{
				Name:        "ip6",
				Usage:       "IPvr redis key for IPv6 location",
				Value:       "",
				Destination: &args.keyIPv6,
			},
			cli.StringFlag{
				Name:        "ip6i",
				Usage:       "IPv6 info redis key for IPv6 location",
				Value:       "",
				Destination: &args.keyIPv6Info,
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

	if args.host == "" || args.port == 0 {
		return fmt.Errorf("Specify redis host and port")
	}

	if args.keyIPv4 == "" || args.keyIPv6 == "" || args.keyIPv6Info == "" {
		return fmt.Errorf("Specify the following flags: ip4 or ip6 and ip6i")
	}

	if args.blocks == "" {
		return fmt.Errorf("Specify one of following flags: import-blocks")
	}

	client := geo.NewGeoClient(args.host, args.port, args.password, args.db, args.keyIPv4, args.keyIPv6, args.keyIPv6Info)

	if args.blocks != "" {
		err := client.ImportBlocks(args.blocks)
		if err != nil {
			return err
		}
	}

	return nil
}
