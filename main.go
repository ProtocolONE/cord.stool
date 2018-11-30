// To install goversioninfo run:
// go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo

package main

import (
	"log"
	"os"

	"cord.stool/cmd"
	"cord.stool/context"
	// "github.com/gosuri/uiprogress"
	// "cord.stool/updater"
)

func main() {

	// uiprogress.Start()            // start rendering
	// bar := uiprogress.AddBar(100) // Add a new bar

	// // optionally, append and prepend completion and elapsed time
	// bar.AppendCompleted()
	// bar.PrependElapsed()

	// r,_ := updater.EnumFilesRecursive(`E:\Prog\Go\stool.fixture\src\test\dst`, make(chan struct{}))

	// for _ = range r {
	// 	bar.Incr();
	// }

	// return
	ctx := context.NewContext()
	cmd.RegisterCmdCommands(ctx)
	err := ctx.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
