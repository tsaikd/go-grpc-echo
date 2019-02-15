package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	throttle "github.com/boz/go-throttle"
	"github.com/tsaikd/KDGoLib/errutil"
	"github.com/tsaikd/go-grpc-echo/logger"
	pb "github.com/tsaikd/go-grpc-echo/pb"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Server echo server side implement struct
type Server struct {
	Hostname string

	grpcServer      *grpc.Server
	httpServer      *http.Server
	subscriberCount int64
}

// init check and prepare used config in Server struct
func (s *Server) init() (err error) {
	if err = s.Close(); err != nil {
		return
	}

	if s.Hostname == "" {
		if s.Hostname, err = os.Hostname(); err != nil {
			return
		}
	}

	return
}

// Listen a http/grpc server
func (s *Server) Listen(
	addr string,
	certPath string,
	keyPath string,
) (err error) {
	logger := logger.Logger()
	if err = s.init(); err != nil {
		return
	}

	opts := []grpc.ServerOption{}
	tlsFlag := false

	if certPath != "" && keyPath != "" {
		creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
		if err != nil {
			return err
		}
		opts = append(opts, grpc.Creds(creds))
		tlsFlag = true
	}

	s.grpcServer = grpc.NewServer(opts...)
	pb.RegisterEchoServer(s.grpcServer, s)

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", s.httpEchoPing)

	rootMux := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.ProtoMajor == 2 && strings.Contains(req.Header.Get("Content-Type"), "application/grpc") {
			s.grpcServer.ServeHTTP(w, req)
		} else {
			httpMux.ServeHTTP(w, req)
		}
	})

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(rootMux, &http2.Server{}),
	}

	logger.Printf("grpc listen address: %q tls: %v", addr, tlsFlag)
	if tlsFlag {
		return s.httpServer.ListenAndServeTLS(certPath, keyPath)
	}
	return s.httpServer.ListenAndServe()
}

// Close server
func (s *Server) Close() (err error) {
	if s.grpcServer != nil {
		s.grpcServer.Stop()
		s.grpcServer = nil
	}
	if s.httpServer != nil {
		err = s.httpServer.Close()
		s.httpServer = nil
	}
	return
}

func (s *Server) httpEchoPing(w http.ResponseWriter, req *http.Request) {
	header, err := json.Marshal(req.Header)
	errutil.Trace(err)

	resp := &pb.Pong{
		Timestamp: time.Now().Unix(),
		Hostname:  s.Hostname,
		Header:    string(header),
	}
	logger.Logger().Printf("Receive http %d.%d %s %q , response: %+q", req.ProtoMajor, req.ProtoMinor, req.Method, req.RequestURI, resp)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(resp.String() + "\n"))
	errutil.Trace(err)
}

// Send return pong
func (s *Server) Send(ctx context.Context, ping *pb.Ping) (*pb.Pong, error) {
	header := ""
	if headerMD, ok := metadata.FromIncomingContext(ctx); ok {
		if headerJSON, err := json.Marshal(headerMD); err == nil {
			header = string(headerJSON)
		}
	}

	resp := &pb.Pong{
		Message:   ping.Message,
		Timestamp: time.Now().Unix(),
		Hostname:  s.Hostname,
		Header:    header,
	}
	logger.Logger().Printf("Receive grpc %+q , response: %+q", ping, resp)
	return resp, nil
}

// Subscribe response message for every second
func (s *Server) Subscribe(ping *pb.Ping, stream pb.Echo_SubscribeServer) (err error) {
	ctx := stream.Context()
	s.changeSubscriberCount(1)
	defer s.changeSubscriberCount(-1)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case timestamp := <-ticker.C:
			err = stream.Send(&pb.Pong{
				Message:   ping.Message,
				Timestamp: timestamp.Unix(),
			})
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Server) changeSubscriberCount(delta int64) {
	atomic.AddInt64(&s.subscriberCount, delta)
	throttleLog(func() {
		logger.Logger().Printf("Subscriber: %d", s.subscriberCount)
	})
}

var throttleFunc = func() {}
var throttleInstance = throttle.ThrottleFunc(5*time.Second, true, func() {
	throttleFunc()
})

func throttleLog(f func()) {
	throttleFunc = f
	throttleInstance.Trigger()
}
