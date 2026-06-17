GO = CGO_ENABLED=0 GO111MODULE=on GOPROXY=https://goproxy.cn,direct go
BUILD_DATE := $(shell date '+%Y-%m-%d %H:%M:%S')
GIT_BRANCH := $(shell git symbolic-ref --short -q HEAD)
GIT_COMMIT_HASH := $(shell git rev-parse HEAD|cut -c 1-8)
GO_FLAGS := -v -ldflags="-X 'github.com/gzzyj/g/build.Built=$(BUILD_DATE)' -X 'github.com/gzzyj/g/build.GitCommit=$(GIT_COMMIT_HASH)' -X 'github.com/gzzyj/g/build.GitBranch=$(GIT_BRANCH)'"
MODULE =github.com/gzzyj/g

BUILD_PATH=${MODULE}/cmd

all: install lint test clean

build:
	mkdir -p bin
	$(GO) build $(GO_FLAGS)  -o ./bin/g  ${BUILD_PATH}

install: build
	$(GO) install $(GO_FLAGS)

build-all: build-linux build-darwin build-windows build-freebsd

build-linux: build-linux-386 build-linux-amd64 build-linux-arm build-linux-arm64 build-linux-s390x build-linux-riscv64
build-linux-386:
	GOOS=linux GOARCH=386 $(GO) build $(GO_FLAGS) -o bin/linux-386/g ${BUILD_PATH}
build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GO) build $(GO_FLAGS) -o bin/linux-amd64/g ${BUILD_PATH}
build-linux-arm:
	GOOS=linux GOARCH=arm $(GO) build $(GO_FLAGS) -o bin/linux-arm/g ${BUILD_PATH}
build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GO) build $(GO_FLAGS) -o bin/linux-arm64/g ${BUILD_PATH}
build-linux-s390x:
	GOOS=linux GOARCH=s390x $(GO) build $(GO_FLAGS) -o  bin/linux-s390x/g ${BUILD_PATH}
build-linux-riscv64:
	GOOS=linux GOARCH=riscv64 $(GO) build $(GO_FLAGS) -o  bin/linux-riscv64/g ${BUILD_PATH}


build-darwin: build-darwin-amd64 build-darwin-arm64
build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(GO) build $(GO_FLAGS) -o bin/darwin-amd64/g ${BUILD_PATH}
build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GO) build $(GO_FLAGS) -o bin/darwin-arm64/g ${BUILD_PATH}


build-windows: build-windows-386 build-windows-amd64 build-windows-arm build-windows-arm64
build-windows-386:
	GOOS=windows GOARCH=386 $(GO) build $(GO_FLAGS) -o bin/windows-386/g.exe ${BUILD_PATH}
build-windows-amd64:
	GOOS=windows GOARCH=amd64 $(GO) build $(GO_FLAGS) -o bin/windows-amd64/g.exe  ${BUILD_PATH}
build-windows-arm64:
	GOOS=windows GOARCH=arm64 $(GO) build $(GO_FLAGS) -o bin/windows-arm64/g.exe ${BUILD_PATH}


build-freebsd: build-freebsd-386 build-freebsd-amd64 build-freebsd-arm build-freebsd-arm64 build-freebsd-riscv64
build-freebsd-386:
	GOOS=freebsd GOARCH=386 $(GO) build $(GO_FLAGS) -o bin/freebsd-386/g ${BUILD_PATH}
build-freebsd-amd64:
	GOOS=freebsd GOARCH=amd64 $(GO) build $(GO_FLAGS) -o bin/freebsd-amd64/g ${BUILD_PATH}
build-freebsd-arm:
	GOOS=freebsd GOARCH=arm $(GO) build $(GO_FLAGS) -o bin/freebsd-arm/g ${BUILD_PATH}
build-freebsd-arm64:
	GOOS=freebsd GOARCH=arm64 $(GO) build $(GO_FLAGS) -o bin/freebsd-arm64/g ${BUILD_PATH}
build-freebsd-riscv64:
	GOOS=freebsd GOARCH=riscv64 $(GO) build $(GO_FLAGS) -o  bin/freebsd-riscv64/g ${BUILD_PATH}

package:
	bash ./package.sh

install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

lint:
	# Please make sure you are using the latest go version before executing lint
	@go version
	go vet ./...
	golangci-lint run ./...
	staticcheck ./...
	gosec -exclude=G107,G204,G304,G401,G505 -quiet ./...

test:
	go test -v -gcflags=all=-l ./...

test-coverage:
	go test -gcflags=all=-l -race -coverprofile=coverage.txt -covermode=atomic ./...

view-coverage: test-coverage
	go tool cover -html=coverage.txt
	rm -f coverage.txt

clean:
	$(GO) clean -x
	rm -f sha256sum.txt
	rm -rf bin
	rm -f coverage.txt

upgrade-deps:
	go get -u -v github.com/urfave/cli/v2@latest
	go get -u -v github.com/Masterminds/semver/v3@latest
	go get -u -v github.com/PuerkitoBio/goquery@latest
	go get -u -v github.com/mholt/archiver/v3@latest
	go get -u -v github.com/schollz/progressbar/v3@latest
	go get -u -v github.com/daviddengcn/go-colortext@latest
	go get -u -v github.com/fatih/color@latest
	go get -u -v github.com/k0kubun/go-ansi@latest
	go get -u -v github.com/agiledragon/gomonkey/v2@latest
	go get -u -v github.com/stretchr/testify@latest
	go get -u -v golang.org/x/text@latest

mcp-inspector: build
	npx @modelcontextprotocol/inspector ./bin/g mcp

.PHONY: all build install install-tools lint test test-coverage view-coverage addlicense package clean upgrade-deps mcp-inspector build-linux build-darwin build-windows build-linux-386 build-linux-amd64 build-linux-arm build-linux-arm64 build-linux-s390x build-linux-riscv64 build-darwin-amd64 build-darwin-arm64 build-windows-386 build-windows-amd64 build-windows-arm build-windows-arm64 build-freebsd build-freebsd-386 build-freebsd-amd64 build-freebsd-arm build-freebsd-arm64 build-freebsd-riscv64
