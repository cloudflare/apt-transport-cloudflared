
IMPORT_PATH := github.com/cloudflare/apt-transport-cloudflared
INSTALL_DIR := /usr/lib/apt/methods/

.PHONEY: all
all: cfd+https test

.PHONEY: clean
clean:
	go clean

.PHONEY: vet
vet:
	@echo "Vetting code with go vet"
	@tput setaf 1
	@go vet ${IMPORT_PATH}/cmd/cfd || tput sgr0
	@tput sgr0

.PHONEY: lint
lint:
	@echo "Linting source with golint"
	@tput setaf 1
	@golint ${IMPORT_PATH}/cmd/cfd || tput sgr0
	@tput sgr0

.PHONEY: cyclo
cyclo:
	@echo "Checking for large functions with gocyclo"
	@tput setaf 1
	@gocyclo -over 10 ./cmd/cfd || tput sgr0
	@tput sgr0

.PHONEY: check
check: vet lint cyclo

.PHONEY: cfd+https
cfd+https: bin/cfd+https

.PHONEY: test
test:
	go test -coverprofile=cover.out -test.v ${IMPORT_PATH}/cmd/cfd

.PHONEY: fmt
fmt:
	gofmt -s -w cmd/cfd/*.go

bin/cfd+https: cmd/cfd/*.go
	go build -o bin/cfd+https ${IMPORT_PATH}/cmd/cfd


