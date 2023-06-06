package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Clean up -source SRCDIR.
func clean_source(srcdir string) string {
	src, err := filepath.Abs(srcdir)
	if (err != nil) {
		panic(err)
	}
	return src
}

// Clean up -destination PKGDIR
func clean_destination(pkgdir string) string {
	dest, err := filepath.Abs(pkgdir)
	if (err != nil) {
		panic(err)
	}
	return dest
}

// Clean up -repository CONNECTION
func clean_repository(connection string) string {
	repo := strings.TrimSpace(connection)

	// To obtain a directory listing, the connection string must end in a
	// forward slash.
	if (strings.HasSuffix(repo, "/") == false) {
		repo += "/"
	}

	pattern, err := regexp.Compile(`^(([A-Za-z0-9][A-Za-z0-9._-]*@)?([A-Za-z0-9._-]+):)?(/[A-Za-z0-9._-]+)+/$`)
	if (err != nil) {
		panic(err)
	}
	match := pattern.FindStringSubmatch(repo)
	if (match == nil) || (match[4] == "") {
		panic(fmt.Sprintf("Connection string %s seems invalid", repo))
	}

	return repo
}

// Clean up -arch ARCH
func clean_architecture(arch, repo string) string {
	if (arch == "amd64" ) || (arch == "arm64") {
		return arch
	}

	if (strings.Contains(repo, "x86_64")) {
		return "amd64"
	} else if (strings.Contains(repo, "aarch64")) {
		return "arm64"
	}

	panic("Not a valid architecture")
}

