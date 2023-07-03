build:
	@go build -o bin/ChlockBane


run: build
	@./bin/ChlockBane

test:
	go test ./... -v -count=1

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	proto/*.proto


.PHONY: proto