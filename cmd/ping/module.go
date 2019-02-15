package ping

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tsaikd/KDGoLib/cliutil/cobrather"
	"github.com/tsaikd/go-grpc-echo/client"
	"github.com/tsaikd/go-grpc-echo/logger"
	"google.golang.org/grpc/metadata"
)

var flagURL = &cobrather.StringFlag{
	Name:    "ping.url",
	Default: "localhost:8080",
	Usage:   "gRPC server URL",
}

var flagMessage = &cobrather.StringFlag{
	Name:    "ping.message",
	Default: "",
	Usage:   "ping message",
}

var flagCert = &cobrather.StringFlag{
	Name:    "ping.cert.path",
	Default: "",
	Usage:   "gRPC server tls certificate file path",
}

var flagInsecure = &cobrather.BoolFlag{
	Name:    "ping.insecure-skip-verify",
	Default: false,
	Usage:   "controls whether a client verifies the server's certificate chain and host name",
}

var flagHeaders = &cobrather.StringSliceFlag{
	Name:    "ping.header",
	Default: []string{},
	Usage:   "custom ping header",
}

// Module info of package
var Module = &cobrather.Module{
	Use:   "ping",
	Short: "ping grpc echo server",
	Flags: []cobrather.Flag{
		flagURL,
		flagMessage,
		flagCert,
		flagInsecure,
		flagHeaders,
	},
	RunE: func(ctx context.Context, cmd *cobra.Command, args []string) (err error) {
		logger.DefaultThrottler.SetContext(ctx)
		url := flagURL.String()
		message := flagMessage.String()
		certPath := flagCert.String()
		insecureSkipVerify := flagInsecure.Bool()
		headers := flagHeaders.StringSlice()
		headerMap := metadata.MD{}
		for _, header := range headers {
			idx := strings.Index(header, "=")
			if idx > 0 {
				key := strings.TrimSpace(header[0:idx])
				value := strings.TrimSpace(header[idx+1:])
				headerMap.Append(key, value)
			}
		}
		return client.Ping(ctx, url, message, certPath, insecureSkipVerify, headerMap)
	},
}
