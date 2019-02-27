
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
	go test ${IMPORT_PATH}/cmd/cfd

bin/cfd+https:
	go build -o bin/cfd+https ${IMPORT_PATH}/cmd/cfd


