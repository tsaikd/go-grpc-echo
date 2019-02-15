# go-grpc-echo
golang grpc echo example

* enable http2 verbose log
```
GODEBUG="http2debug=1"
```

* run server
```
go-grpc-echo server
```

* run echo client
```
go-grpc-echo ping --ping.url localhost:8080
```

* run grpcurl
```
grpcurl -plaintext -proto pb/echo.proto localhost:8080 go_grpc_echo_pb.Echo.Send
```
