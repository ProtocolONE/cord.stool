package cmd

import (
	"cord.stool/cmd/create"
	"cord.stool/cmd/diff"
	"cord.stool/cmd/push"
	"cord.stool/cmd/service"
	"cord.stool/cmd/torrent"
	"cord.stool/cmd/upgrade"
	"cord.stool/cmd/games"
	"cord.stool/context"
)

func RegisterCmdCommands(ctx *context.StoolContext) {
	create.Register(ctx)
	push.Register(ctx)
	torrent.Register(ctx)
	diff.Register(ctx)
	upgrade.Register(ctx)
	service.Register(ctx)
	games.Register(ctx)
}
