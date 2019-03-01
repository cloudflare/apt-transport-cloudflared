
IMPORT_PATH := github.com/cloudflare/apt-transport-cloudflared
INSTALL_DIR := /usr/lib/apt/methods/

.PHONEY: all
all: cfd+https test

.PHONEY: clean
clean:
	go clean

.PHONEY: cfd+https
cfd+https: bin/cfd+https

.PHONEY: test
test:
	go test -coverprofile=cover.out -test.v ${IMPORT_PATH}/cmd/cfd

.PHONEY: fmt
fmt:
	gofmt -w cmd/cfd/*.go

bin/cfd+https: cmd/cfd/*.go
	go build -o bin/cfd+https ${IMPORT_PATH}/cmd/cfd


