.PHONY: build clean gosumgen

build: gosumgen
	go build -o bin/main cmd/main.go

clean:
	rm go.sum

gosumgen: clean
	go mod tidy

go: clean build
	go run cmd/main.go

test:
	go test ./...