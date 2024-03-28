go.mod:
	go mod init git.dominic-ricottone.com/~dricottone/simple-builder
	go get github.com/docker/docker/client
	go get github.com/docker/docker/api/types/container
	go get github.com/docker/docker/api/types/mount
	go get github.com/opencontainers/image-spec/specs-go/v1

simple-builder: go.mod *.go
	go build .

build: simple-builder

clean:
	rm -f go.mod go.sum simple-builder

uninstall:
	rm -f ~/.local/bin/simple-builder

PWD=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))
install:
	ln -s $(PWD)simple-builder ~/.local/bin/simple-builder

.PHONY: build clean install
