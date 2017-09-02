## Setup environment for grpc
```
wget https://github.com/google/protobuf/releases/download/v3.3.0/protobuf-cpp-3.3.0.tar.gz
tar xvfz protobuf-cpp-3.3.0.tar.gz
cd protobuf-3.3.0/
./configure && make && make install
go get -v google.golang.org/grpc
go get -a github.com/golang/protobuf/protoc-gen-go

```

## Generate the source file from proto IDL
```
# make sure $GOBIN is added in the $PATH
protoc --go_out=plugins=grpc:. helloworld.proto
```

## Build
```
go build client.go
go build server.go
```