// To install goversioninfo run:
// go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -64 -icon=launcher.ico

package main

import (
	"log"
	"os"

	"cord.stool/cmd"
	"cord.stool/context"
)

var version = "develop"

func main() {

	ctx := context.NewContext(version)
	cmd.RegisterCmdCommands(ctx)
	err := ctx.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
