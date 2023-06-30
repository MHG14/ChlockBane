build:
	@go build -o bin/ChlockBane


run: build
	@./bin/docker

test:
	go test ./... -v -count=1