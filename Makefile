BIN_DIR?=/usr/local/bin

INSTALL_FLAGS?=
INSTALL_FLAGS+=--mode=755

go.mod:
	go mod init git.sr.ht/~dricottone/simple-builder
	go get github.com/docker/docker/client
	go get github.com/docker/docker/api/types/container
	go get github.com/docker/docker/api/types/mount
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
