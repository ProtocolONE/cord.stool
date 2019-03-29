package geo

import (
	"fmt"

	"cord.stool/context"
	"cord.stool/geo"

	"github.com/urfave/cli"
)

var args = struct {
	host      string
	port      string
	blocks    string
	locations string
	ip        string
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
				Name:        "import-blocks, ib",
				Usage:       "Imports the location blocks file, import-blocks <file>.",
				Value:       "",
				Destination: &args.blocks,
			},
			cli.StringFlag{
				Name:        "import-locations, il",
				Usage:       "Imports the location details file, import-locations <file>",
				Value:       "",
				Destination: &args.locations,
			},
			cli.StringFlag{
				Name:        "lookup, l",
				Usage:       "Looks up the geo data for the IP, lookup <ip>",
				Value:       "",
				Destination: &args.ip,
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

	if args.blocks == "" && args.locations == "" && args.ip == "" {
		return fmt.Errorf("Specify one of following flags: import-blocks, import-locations, lookup")
	}

	client := geo.NewGeoClient(args.host, args.port)

	if args.blocks != "" {
		err := client.ImportBlocks(args.blocks)
		if err != nil {
			return err
		}
	}

	if args.locations != "" {
		err := client.ImportLocations(args.locations)
		if err != nil {
			return err
		}
	}

	if args.ip != "" {
		loc, err := client.LookupLocation(args.ip)
		if err != nil {
			return err
		}
		fmt.Println(loc)
	}

	return nil
}
