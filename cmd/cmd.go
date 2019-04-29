package cmd

import (
	"cord.stool/cmd/branch"
	"cord.stool/cmd/create"
	"cord.stool/cmd/diff"
	"cord.stool/cmd/games"
	"cord.stool/cmd/push"
	"cord.stool/cmd/service"
	"cord.stool/cmd/torrent"
	"cord.stool/cmd/upgrade"
	"cord.stool/cmd/build"
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
	branch.Register(ctx)
	build.Register(ctx)
}
