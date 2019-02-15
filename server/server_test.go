package server

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/go-grpc-echo/client"
	"golang.org/x/sync/errgroup"
)

func getAddr() (addr string) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}
	addr = lis.Addr().String()
	if err = lis.Close(); err != nil {
		panic(err)
	}
	return
}

func Test_http1_without_tls(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	server := &Server{}
	addr := getAddr()
	certPath := ""
	keyPath := ""

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return server.Listen(addr, certPath, keyPath)
	})
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			errutil.Trace(server.Close())
			cancel()
		}
		return nil
	})

	time.Sleep(100 * time.Millisecond)

	client := http.Client{}
	resp, err := client.Get("http://" + addr)
	require.NoError(err)
	require.EqualValues(http.StatusOK, resp.StatusCode)

	errutil.Trace(server.Close())
	require.EqualError(eg.Wait(), http.ErrServerClosed.Error())
}

func Test_http1_with_tls(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	server := &Server{}
	addr := getAddr()
	certPath := "test/cert.pem"
	keyPath := "test/key.pem"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return server.Listen(addr, certPath, keyPath)
	})
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			errutil.Trace(server.Close())
			cancel()
		}
		return nil
	})

	time.Sleep(100 * time.Millisecond)

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	resp, err := client.Get("https://" + addr)
	require.NoError(err)
	require.EqualValues(http.StatusOK, resp.StatusCode)

	errutil.Trace(server.Close())
	require.EqualError(eg.Wait(), http.ErrServerClosed.Error())
}

func Test_grpc_without_tls(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	server := &Server{}
	addr := getAddr()
	certPath := ""
	keyPath := ""

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return server.Listen(addr, certPath, keyPath)
	})
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			errutil.Trace(server.Close())
			cancel()
		}
		return nil
	})

	time.Sleep(100 * time.Millisecond)

	err := client.Ping(ctx, addr, "", certPath, false, nil)
	require.NoError(err)
}

func Test_grpc_with_tls(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	server := &Server{}
	addr := getAddr()
	certPath := "test/cert.pem"
	keyPath := "test/key.pem"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return server.Listen(addr, certPath, keyPath)
	})
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			errutil.Trace(server.Close())
			cancel()
		}
		return nil
	})

	time.Sleep(100 * time.Millisecond)

	err := client.Ping(ctx, addr, "", certPath, true, nil)
	require.NoError(err)
}
