package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"time"

	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/go-grpc-echo/logger"
	pb "github.com/tsaikd/go-grpc-echo/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Ping send Ping message to grpc server
func Ping(ctx context.Context, url string, message string, certPath string, insecureSkipVerify bool, headerMap metadata.MD) (err error) {
	opts := []grpc.DialOption{}

	var tlsConfig *tls.Config
	if certPath != "" {
		certFile, err := ioutil.ReadFile(certPath)
		if err != nil {
			return err
		}
		certPool := x509.NewCertPool()
		if certPool.AppendCertsFromPEM(certFile) {
			tlsConfig = &tls.Config{RootCAs: certPool, InsecureSkipVerify: insecureSkipVerify}
		}
	}

	if insecureSkipVerify && tlsConfig == nil {
		tlsConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	}

	if tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(url, opts...)
	if err != nil {
		return
	}
	defer func() {
		errutil.Trace(conn.Close())
	}()

	ctx = metadata.NewOutgoingContext(ctx, headerMap)
	msg := &pb.Ping{Message: message}
	client := pb.NewEchoClient(conn)
	pong, err := client.Send(ctx, msg)
	if err != nil {
		return
	}
	logger.TrottlePrintf("Send %+q to %q (tls: %v), received %+q", msg, url, tlsConfig != nil, pong)
	return
}

// Subscribe stream from grpc server
func Subscribe(ctx context.Context, url string, message string, duration time.Duration) (err error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return
	}
	defer func() {
		errutil.Trace(conn.Close())
	}()

	msg := &pb.Ping{Message: message}
	client := pb.NewEchoClient(conn)
	sub, err := client.Subscribe(ctx, msg)
	if err != nil {
		return
	}
	defer func() {
		errutil.Trace(sub.CloseSend())
	}()

	timer := time.NewTimer(duration)
	for {
		select {
		case <-ctx.Done():
			return
		case <-sub.Context().Done():
			return
		case <-timer.C:
			return
		default:
		}

		pong, err := sub.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		logger.TrottlePrintf("Send %+q to %q , received %+q", msg, url, pong)
	}
}
