BINARY  := sni-firewall
PKG     := ./cmd/sni-firewall
DIST    := dist

.PHONY: all build build-linux test lint vet fmt clean

all: lint test build

build:
	go build -trimpath -o $(DIST)/$(BINARY) $(PKG)

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o $(DIST)/$(BINARY)-linux-amd64 $(PKG)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -o $(DIST)/$(BINARY)-linux-arm64 $(PKG)

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

clean:
	rm -rf $(DIST)
