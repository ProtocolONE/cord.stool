//go:generate goversioninfo

package main

import (
	"log"
	"os"

	"cord.stool/cmd"
	"cord.stool/context"
)

func main() {
	ctx := context.NewContext()
	cmd.RegisterCmdCommands(ctx)
	err := ctx.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
