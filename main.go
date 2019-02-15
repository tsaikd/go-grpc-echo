package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tsaikd/KDGoLib/cliutil/cobrather"
	"github.com/tsaikd/go-grpc-echo/cmd/ping"
	"github.com/tsaikd/go-grpc-echo/cmd/server"
	"github.com/tsaikd/go-grpc-echo/cmd/subscribe"
)

// Module info of package
var Module = &cobrather.Module{
	Use:   "go-grpc-echo",
	Short: "grpc echo client/server in golang",
	Commands: []*cobrather.Module{
		ping.Module,
		server.Module,
		subscribe.Module,
		cobrather.VersionModule,
	},
	RunE: func(ctx context.Context, cmd *cobra.Command, args []string) (err error) {
		if len(args) < 1 {
			return cmd.Help()
		}
		return
	},
}

func main() {
	Module.MustMainRun()
}
