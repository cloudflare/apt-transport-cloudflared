VERSION     := $(shell git describe --tags --always --dirty=-dev)
NAME        := apt-transport-cloudflared
BUILD_PATH  := dist
DEB_NAME    := apt-transport-cloudflared_${VERSION}_amd64.deb
FPM_ARGS    := --provides apt-transport-cloudflared -v ${VERSION}

.PHONY: all
all: cfd+https

.PHONY: clean
clean:
	go clean
	rm -rf ./bin/ *.deb

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
	go test -coverprofile=cover.out -test.v ./apt ./apt/access

.PHONY: build
build: check bin/cfd+https

.PHONY: fmt
fmt:
	gofmt -s -w cmd/cfd/*.go apt/*.go apt/**/*.go
	goimports -w cmd/cfd/*.go apt/*.go apt/**/*.go

bin/cfd+https: cmd/cfd/*.go apt/*.go apt/**/*.go
	go build -o bin/cfd+https ./cmd/cfd

.PHONY: package
package: ${DEB_NAME}

${DEB_NAME}: bin/cfd+https
	mkdir -p ${BUILD_PATH}/usr/lib/apt/methods/
	cp bin/cfd+https ${BUILD_PATH}/usr/lib/apt/methods/cfd+https
	fpm -t deb --deb-user root --deb-group root -s dir ${FPM_ARGS} -n ${NAME} -C ${BUILD_PATH} \
		--deb-no-default-config-files
