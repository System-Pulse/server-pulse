NAME=server-pulse
VERSION=$(shell cat VERSION)
BUILD=$(shell git rev-parse --short HEAD)
LD_FLAGS="-w -X main.version=$(VERSION) -X main.build=$(BUILD)"

clean:
	rm -rf _build/ release/

build:
	go mod download
	CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o server-pulse

build-all:
	mkdir -p _build
	GOOS=darwin  GOARCH=amd64   CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/server-pulse-$(VERSION)-darwin-amd64
	GOOS=linux   GOARCH=amd64   CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/server-pulse-$(VERSION)-linux-amd64
	GOOS=linux   GOARCH=arm     CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/server-pulse-$(VERSION)-linux-arm
	GOOS=linux   GOARCH=arm64   CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/server-pulse-$(VERSION)-linux-arm64
	GOOS=linux   GOARCH=ppc64le CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/server-pulse-$(VERSION)-linux-ppc64le
	GOOS=windows GOARCH=amd64   CGO_ENABLED=0 go build -tags release -ldflags $(LD_FLAGS) -o _build/server-pulse-$(VERSION)-windows-amd64
	cd _build; sha256sum * > sha256sums.txt

run-dev:
	rm -f server-pulse.sock server-pulse
	go build -ldflags $(LD_FLAGS) -o server-pulse
	server-pulse_DEBUG=1 ./server-pulse

image:
	docker build -t server-pulse -f Dockerfile .

release:
	mkdir -p release
	cp _build/* release
	cd release; sha256sum --quiet --check sha256sums.txt && \
	gh release create v$(VERSION) -d -t v$(VERSION) *

.PHONY: build
