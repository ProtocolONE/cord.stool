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
	
	//"github.com/akamai/netstoragekit-golang"
	//"fmt"
)

var version = "develop"

func main() {

	/*nsHostname := "akamai.cdn.protocol.one"
	nsKeyname  := "p1upload"
	nsKey := "bSOmA4aHPYfWLd2uFgzohewFFhnE3rN7KaPfpv0z"
	nsCpcode := "360949"
  
	ns := netstorage.NewNetstorage(nsHostname, nsKeyname, nsKey, false)
  
	localSource := "text.txt"
	nsDestination := fmt.Sprintf("/%s/hello.txt", nsCpcode) // or "/%s/" is same. 
  
	res, body, err1 := ns.Upload(localSource, nsDestination)
	if err1 != nil {
		// Do something
	}
  
	if res.StatusCode == 200 {
		fmt.Printf(body)
	}	

	res, body, err1 = ns.Download(nsDestination, "D:\\Projects\\Syncopate\\sources\\ProtocolONE\\cord.stool\\text2.txt")
	if err1 != nil {
		// Do something
	}

	if res.StatusCode == 200 {
		fmt.Printf(body)
	}*/

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
	ctx := context.NewContext(version)
	cmd.RegisterCmdCommands(ctx)
	err := ctx.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
