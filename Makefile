.PHONY: build build-windows build-mac build-linux build-all clean

BINARY = sharedlink
BIN_DIR = bin

build:
	go build -o $(BIN_DIR)/$(BINARY) ./cmd/$(BINARY)

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY)-windows-amd64.exe ./cmd/$(BINARY)

build-mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY)-darwin-amd64 ./cmd/$(BINARY)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/$(BINARY)-darwin-arm64 ./cmd/$(BINARY)

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY)-linux-amd64 ./cmd/$(BINARY)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(BIN_DIR)/$(BINARY)-linux-arm64 ./cmd/$(BINARY)

build-all: build-windows build-mac build-linux

clean:
	rm -rf $(BIN_DIR)
