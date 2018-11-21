package push

import (
	"fmt"

	"cord.stool/context"
	"github.com/urfave/cli"
)

func Register(ctx *context.StoolContext) {
	cmd := cli.Command{
		Name:        "push",
		Usage:       "Upload update",
		Description: "Upload update app bundle to one of servers",

		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "ftp",
				Value: "",
			},
		},
		Action: func(c *cli.Context) error {
			return do(ctx, c)
		},
	}
	ctx.App.Commands = append(ctx.App.Commands, cmd)
}

func do(ctx *context.StoolContext, c *cli.Context) error {
	fmt.Printf("push update\n")
	if ctx.Verbose {
		fmt.Println("verbose show")
	}

	//fmt.Println(ctx.App, c)
	fmt.Println(c.Args())
	return nil
}
