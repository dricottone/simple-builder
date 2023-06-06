package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path"
	"regexp"
)

var (
	pattern_rsync_stdout = regexp.MustCompile(`^[drwx-]{10} *[0-9,]+ *[0-9]{4}/[0-9]{2}/[0-9]{2} *[0-9]{2}:[0-9]{2}:[0-9]{2} ([A-Za-z0-9._-]+) *$`)
)

// Fetch a listing from a directory that is serving as a package repository.
func fetch_repository_listing(remote_dir string) ([]Package, error) {
	debug(fmt.Sprintf("DEBUG-RSYNC:rsync --list-only %s", remote_dir))
	cmd := exec.Command("rsync", "--list-only", remote_dir)
	stdout, err := cmd.StdoutPipe()
	if (err != nil) {
		return nil, err
	}
	err = cmd.Start()
	if (err != nil) {
		return nil, err
	}

	pkgs, err := parse_rsync_stdout(stdout)
	if (err != nil) {
		return nil, err
	}

	err = cmd.Wait()
	if (err != nil) {
		return nil, err
	}

	if (len(pkgs) == 0) {
		return nil, errors.New("No packages found")
	}

	uniq := []Package{}
	for _, pkg := range pkgs {
		i := find_package(&uniq, pkg.Name)
		if (i == -1) {
			uniq = append(uniq, pkg)
		} else {
			diff, err := compare_versions(uniq[i].Version, pkg.Version)
			if (err != nil) {
				return nil, err
			}
			if (diff == 1) {
				uniq[i].Version = pkg.Version
			}
		}
	}

	return uniq, nil
}

// Parse `rsync(1)` output.
func parse_rsync_stdout(stdout io.Reader) ([]Package, error) {
	packages := []Package{}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		pkgs, err := parse_rsync_line(line)
		if (err != nil) {
			return nil, err
		}

		packages = append(packages, pkgs...)
	}

	err := scanner.Err()
	if (err != nil) {
		return nil, err
	}

	return packages, nil
}

// Parse a line of `rsync(1)` output into 0 or 1 packages.
func parse_rsync_line(line string) ([]Package, error) {
	debug(fmt.Sprintf("DEBUG-RSYNC:%s", line))

	match := pattern_rsync_stdout.FindStringSubmatch(line)
	if (match != nil) {
		name := match[1]

		// Return an empty slice. Not an error, but also not a package.
		if (name == ".") || (name == "APKINDEX.tar.gz") {
			return []Package{}, nil
		}

		// Repackage return value into slice.
		pkg, err := find_apk(name)
		if (err != nil) {
			return nil, err
		} else {
			return []Package{pkg}, nil
		}
	}

	return nil, errors.New("Failed to parse line of rsync stdout")
}

// Push a built package and an updated APKINDEX to a package repository.
func push_package(pkg Package, local_dir, remote_dir string) error {
	local_name := path.Join(local_dir, expected_apk(pkg))

	debug(fmt.Sprintf("DEBUG-RSYNC:rsync %s %s", local_name, remote_dir))
	cmd := exec.Command("rsync", local_name, remote_dir)
	err := cmd.Run()
	if (err != nil) {
		return err
	}

	local_name = path.Join(local_dir, "APKINDEX.tar.gz")

	debug(fmt.Sprintf("DEBUG-RSYNC:rsync %s %s", local_name, remote_dir))
	cmd = exec.Command("rsync", local_name, remote_dir)
	err = cmd.Run()
	if (err != nil) {
		return err
	}

	return nil
}

