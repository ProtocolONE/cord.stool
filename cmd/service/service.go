package service

import (
	"cord.stool/context"
	"cord.stool/service"
	"github.com/urfave/cli"
)

var args = struct {
	port uint
}{
	port: 8080,
}


func Register(ctx *context.StoolContext) {
	cmd := cli.Command{
		Name:        	"service",
		ShortName:		"s",
		Usage:       	"Service mode",
		Description:	"Run the application as service",

		Flags: []cli.Flag{
			cli.UintFlag{
				Name:        "port, p",
				Usage:       "Port number",
				Value:       8080,
				Destination: &args.port,
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func do(ctx *context.StoolContext, c *cli.Context) error {
	return service.Start(args.port);
}
