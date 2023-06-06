package main

import (
	"errors"
	"fmt"
	"os"
	"path"
)

// Walk a local directory to identify package sources.
//
// Package sources should be organized like:
// ```
// root
//  - package1
//     - APKBUILD
//     - ...
//  - ...
// ```
// Any files not matching `APKBUILD` are ignored.
// Any directories not containing an `APKBUILD` file are ignored.
// Any files directly under the root are ignored.
func walk_package_sources(root string) ([]Package, error) {
	packages := []Package{}

	members, err := os.ReadDir(root)
	if (err != nil) {
		return nil, err
	}

	for _, member := range members {
		if (member.IsDir() == true) {
			name := member.Name()
			pkg := new_package(name)

			err = find_apkbuild(&pkg, path.Join(root, name))
			if (err != nil) {
				debug(fmt.Sprintf("DEBUG-PKGSRC:%s", err))
			} else {
				packages = append(packages, pkg)
				debug(fmt.Sprintf("DEBUG-PKGSRC:Package %s found", name))
			}
		}
	}

	if (len(packages) == 0) {
		return nil, errors.New("No packages found")
	}

	return packages, nil
}

