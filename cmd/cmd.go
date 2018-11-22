package cmd

import (
	"cord.stool/context"
	"cord.stool/cmd/create"
	"cord.stool/cmd/push"
	"cord.stool/cmd/torrent"
)

func RegisterCmdCommands(ctx *context.StoolContext) {
	create.Register(ctx)
	push.Register(ctx)
	torrent.Register(ctx)
}
