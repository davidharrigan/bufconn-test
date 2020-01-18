# generate go files from protobuf files
protoc:
	protoc -I grpc/protos grpc/protos/*.proto --go_out=plugins=grpc:grpc/protos

# run unit tests
.PHONY: test
test:
	go test -v ./...
