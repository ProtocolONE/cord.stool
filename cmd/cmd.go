package cmd

import (
	"cord.stool/context"
	"cord.stool/cmd/create"
	"cord.stool/cmd/push"
	"cord.stool/cmd/torrent"
	"cord.stool/cmd/diff"
	"cord.stool/cmd/upgrade"
)

func RegisterCmdCommands(ctx *context.StoolContext) {
	create.Register(ctx)
	push.Register(ctx)
	torrent.Register(ctx)
	diff.Register(ctx)
	upgrade.Register(ctx)
}
