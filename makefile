gen:
	protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:.
init:
	go mod init go-cancel

server:
	go run server.go

.PHONY: gen init server