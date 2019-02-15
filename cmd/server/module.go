package server

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/tsaikd/KDGoLib/cliutil/cobrather"
	"github.com/tsaikd/go-grpc-echo/logger"
	"github.com/tsaikd/go-grpc-echo/server"
)

var flagAddr = &cobrather.StringFlag{
	Name:    "server.addr",
	Default: ":8080",
	Usage:   "gRPC server listen port",
}

var flagCert = &cobrather.StringFlag{
	Name:    "server.cert.path",
	Default: "",
	Usage:   "gRPC server tls certificate file path",
}

var flagKey = &cobrather.StringFlag{
	Name:    "server.key.path",
	Default: "",
	Usage:   "gRPC server tls key file path",
}

// Module info of package
var Module = &cobrather.Module{
	Use:   "server",
	Short: "grpc echo server",
	Flags: []cobrather.Flag{
		flagAddr,
		flagCert,
		flagKey,
	},
	RunE: func(ctx context.Context, cmd *cobra.Command, args []string) (err error) {
		logger.DefaultThrottler.SetContext(ctx)
		addr := flagAddr.String()
		certPath := flagCert.String()
		keyPath := flagKey.String()
		server := &server.Server{}
		return server.Listen(addr, certPath, keyPath)
	},
}
