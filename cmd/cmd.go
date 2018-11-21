package cmd

import (
	"cord.stool/context"
	"cord.stool/cmd/create"
	"cord.stool/cmd/push"
)

func RegisterCmdCommands(ctx *context.StoolContext) {
	create.Register(ctx)
	push.Register(ctx)
}
