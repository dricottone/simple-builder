BIN_DIR?=/usr/local/bin

INSTALL_FLAGS?=
INSTALL_FLAGS+=--mode=755

go.mod:
	go mod init tangled.org/dominic-ricottone.com/simple-builder
	go get github.com/moby/moby/client
	go get github.com/moby/moby/api/types/container
	go get github.com/moby/moby/api/types/mount
	go get github.com/opencontainers/image-spec/specs-go/v1

simple-builder: go.mod *.go
	go build .

build: simple-builder

clean:
	rm -f go.mod go.sum simple-builder

install:
	install --target-directory=$(BIN_DIR) $(INSTALL_FLAGS) simple-builder

uninstall:
	cd $(BIN_DIR) && rm simple-builder

.PHONY: build clean install uninstall
