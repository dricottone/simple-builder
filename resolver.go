package main

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	pattern_version = regexp.MustCompile(`^([0-9]+)\.([0-9]+)(\.([0-9]+))?(\.([0-9]+))?-r([0-9]+)$`)
)

func find_string(haystack *[]string, needle string) int {
	for i, s := range (*haystack) {
		if (s == needle) {
			return i
		}
	}
	return -1
}

func find_package(packages *[]Package, name string) int {
	for i, p := range (*packages) {
		if (p.Name == name) {
			return i
		}
	}
	return -1
}

// Determine the resolution order for packages in a dependencies chain.
//
// This implementation is derivative of the algorithm described at
// https://www.electricmonk.nl/docs/dependency_resolving_algorithm/dependency_resolving_algorithm.html .
// This document is copyrighted by Ferry Boender (2008-2018) and was first
// published May 25, 2010.
func resolve_dependencies(resolved, pkgs *[]Package, index int, visited *[]string) error {
	pkg := (*pkgs)[index]

	i := find_package(resolved, pkg.Name)
	if (i != -1) {
		return nil
	}

	*visited = append(*visited, pkg.Name)

	for _, dep := range pkg.Dependencies {
		// If dep is resolved, skip
		i = find_package(resolved, dep)
		if (i != -1) {
			continue
		}

		i = find_string(visited, dep)
		if (i != -1) {
			return fmt.Errorf("Circular dependencies in %s and %s", pkg.Name, dep)
		}

		// If dep is not a known package, skip
		i = find_package(pkgs, dep)
		if (i == -1) {
			continue
		}

		err := resolve_dependencies(resolved, pkgs, i, visited)
		if (err != nil) {
			return err
		}
	}

	*resolved = append(*resolved, pkg)
	return nil
}

// Find packages to build.
func find_builds(pkg *Package, repository *[]Package) error {
	i := find_package(repository, (*pkg).Name)

	// Package is new.
	if (i == -1) {
		(*pkg).Build = true
		(*pkg).Message = "new"
		return nil
	}

	ver := (*repository)[i].Version
	diff, err := compare_versions(ver, (*pkg).Version)
	if (err != nil) {
		return err
	}

	// Package is newer in repository. Probably an issue.
	if (diff == -1) {
		(*pkg).Message = fmt.Sprintf("repository has newer %s", ver)
		(*pkg).Error = true
		return nil
	}

	// Package already exists, nothing to do.
	if (diff == 0) {
		return nil
	}

	// Package has an update.
	(*pkg).Build = true
	(*pkg).Message = fmt.Sprintf("update from %s", ver)
	return nil
}

// Find builds that may break other packages that depend on the rebuild.
func find_breaking_builds(pkgs *[]Package) error {
	for _, pkg := range (*pkgs) {
		if (pkg.Build == true) {
			continue
		}

		// If any dependency is marked for build, this package might break
		for _, dep := range pkg.Dependencies {
			i := find_package(pkgs, dep)
			if (i != -1) && ((*pkgs)[i].Build == false) {
				return fmt.Errorf("Package %s depends on updated/new %s but won't be rebuilt", pkg.Name, dep)
			}
		}
	}

	return nil
}

// Parse a version string.
func parse_version_string(version string) ([5]int, error) {
	ver := [5]int{}

	match := pattern_version.FindStringSubmatch(version)
	if (match == nil) {
		return ver, fmt.Errorf("cannot parse %s", version)
	}

	major, err := strconv.Atoi(match[1])
	if (err != nil) {
		return ver, err
	}
	ver[0] = major

	minor, err := strconv.Atoi(match[2])
	if (err != nil) {
		return ver, err
	}
	ver[1] = minor

	sub, err := strconv.Atoi(match[4])
	if (err != nil) && (match[4] != "") {
		return ver, err
	}
	ver[2] = sub

	subsub, err := strconv.Atoi(match[6])
	if (err != nil) && (match[6] != "") {
		return ver, err
	}
	ver[3] = subsub

	release, err := strconv.Atoi(match[7])
	if (err != nil) {
		return ver, err
	}
	ver[4] = release

	debug(fmt.Sprintf("DEBUG-RESOLVER:%s -> %s", version, ver))

	return ver, nil
}

// Compare two version strings.
func compare_versions(remote, local string) (int, error) {
	remote_ver, err := parse_version_string(remote)
	if (err != nil) {
		return 0, err
	}

	local_ver, err := parse_version_string(local)
	if (err != nil) {
		return 0, err
	}

	if (local_ver[0] == remote_ver[0]) {
		if (local_ver[1] == remote_ver[1]) {
			if (local_ver[2] == remote_ver[2]) {
				if (local_ver[3] == remote_ver[3]) {
					if (local_ver[4] == remote_ver[4]) {
						return 0, nil
					} else if (local_ver[4] < remote_ver[4]) {
						return -1, nil
					}
				} else if (local_ver[3] < remote_ver[3]) {
					return -1, nil
				}
			} else if (local_ver[2] < remote_ver[2]) {
				return -1, nil
			}
		} else if (local_ver[1] < remote_ver[1]) {
			return -1, nil
		}
	} else if (local_ver[0] < remote_ver[0]) {
		return -1, nil
	}

	return 1, nil
}

