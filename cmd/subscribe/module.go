package subscribe

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"github.com/tsaikd/KDGoLib/cliutil/cobrather"
	"github.com/tsaikd/go-grpc-echo/client"
	"github.com/tsaikd/go-grpc-echo/logger"
	"golang.org/x/sync/errgroup"
)

var flagURL = &cobrather.StringFlag{
	Name:    "subscribe.url",
	Default: "localhost:8080",
	Usage:   "gRPC server URL",
}

var flagMessage = &cobrather.StringFlag{
	Name:    "subscribe.message",
	Default: "",
	Usage:   "ping message",
}

var flagDuration = &cobrather.StringFlag{
	Name:    "subscribe.duration",
	Default: "10m",
	Usage:   "subscribe duration",
}

var flagConcurrent = &cobrather.Int64Flag{
	Name:    "subscribe.concurrent",
	Default: 1,
	Usage:   "subscribe concurrent connections",
}

// Module info of package
var Module = &cobrather.Module{
	Use:   "subscribe",
	Short: "subscribe grpc echo server",
	Flags: []cobrather.Flag{
		flagURL,
		flagMessage,
		flagDuration,
		flagConcurrent,
	},
	RunE: func(ctx context.Context, cmd *cobra.Command, args []string) (err error) {
		logger.DefaultThrottler.SetContext(ctx)
		duration, err := time.ParseDuration(flagDuration.String())
		if err != nil {
			return
		}
		url := flagURL.String()
		message := flagMessage.String()
		concurrent := flagConcurrent.Int64()
		eg, ctx := errgroup.WithContext(ctx)
		for i := int64(0); i < concurrent; i++ {
			eg.Go(func() error {
				changeSubscriberCount(1)
				defer changeSubscriberCount(-1)
				return client.Subscribe(ctx, url, message, duration)
			})
		}
		return eg.Wait()
	},
}

var currentClientCount int64

func changeSubscriberCount(delta int64) {
	atomic.AddInt64(&currentClientCount, delta)
	logger.TrottlePrintf("Subscriber: %d", currentClientCount)
}
