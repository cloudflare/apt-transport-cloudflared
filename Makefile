IMPORT_PATH := github.com/cloudflare/apt-transport-cloudflared

.PHONY: all
all: cfd+https

.PHONY: clean
clean:
	go clean
	rm -rf ./bin/

.PHONY: vet
vet:
	@./tools/vet.sh ./cmd/cfd ./apt ./apt/exec ./apt/access

.PHONY: check
check: vet
	golangci-lint run

.PHONY: cfd+https
cfd+https: bin/cfd+https

.PHONY: test
test: check
	go test -coverprofile=cover.out -test.v ${IMPORT_PATH}/apt

.PHONY: build
build: check bin/cfd+https

.PHONY: fmt
fmt:
	gofmt -s -w cmd/cfd/*.go apt/*.go apt/**/*.go
	goimports -w cmd/cfd/*.go apt/*.go apt/**/*.go

bin/cfd+https: cmd/cfd/*.go apt/*.go apt/**/*.go
	go build -o bin/cfd+https ${IMPORT_PATH}/cmd/cfd

${DEB_NAME}: clean bin/cfd+https
	mkdir -p ${BUILD_PATH}/usr/lib/apt/methods/
	cp bin/cfd+https ${BUILD_PATH}/usr/lib/apt/methods/cfd+https

