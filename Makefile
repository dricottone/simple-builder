simple-builder:
	go get -u
	go build .

clean:
	rm --force go.sum moby-demo
	rm --force --recursive dir1 dir2

PWD=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))
install:
	ln -s $(PWD)simple-builder ~/.local/bin/simple-builder

.PHONY: clean install
