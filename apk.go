package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	pattern_pkgver = regexp.MustCompile(`^pkgver=(.*)$`)
	pattern_pkgrel = regexp.MustCompile(`^pkgrel=(.*)$`)
	pattern_depends = regexp.MustCompile(`^depends="(.*)"$`)
	pattern_apkname = regexp.MustCompile(`^([A-Za-z0-9._-]+)-([0-9]+\.[0-9]+(\.[0-9]+)?(\.[0-9]+)?-r[0-9]+)\.apk$`)
)

// Scan a directory for an APKBUILD file.
func find_apkbuild(pkg *Package, directory string) error {
	members, err := os.ReadDir(directory)
	if (err != nil) {
		return err
	}

	for _, member := range members {
		if (member.Name() == "APKBUILD") {
			err = parse_apkbuild(pkg, path.Join(directory, member.Name()))
			if (err != nil) {
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("No APKBUILD in %s", pkg.Name)
}

// Parse an APKBUILD file. Given an existing Package, add core information
// (Version, Dependencies) as it is identified.
func parse_apkbuild(pkg *Package, filename string) error {
	pkgver := ""
	pkgrel := ""
	depends := []string{}
	inside_depends := false

	file, err := os.Open(filename)
	if (err != nil) {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		match := pattern_pkgver.FindStringSubmatch(line)
		if (match != nil) {
			pkgver = match[1]
		}
		match = pattern_pkgrel.FindStringSubmatch(line)
		if (match != nil) {
			pkgrel = match[1]
		}
		
		if (inside_depends == true) {
			if (line == "\"") {
				inside_depends = false
			} else {
				depends = append(depends, parse_list(line)...)
			}
		} else if (line == "depends=\"") {
			inside_depends = true
		}

		match = pattern_depends.FindStringSubmatch(line)
		if (match != nil) {
			depends = append(depends, parse_list(match[1])...)
		}
	}

	err = scanner.Err()
	if (err != nil) {
		return err
	}

	if (pkgver == "") || (pkgrel == "") {
		return fmt.Errorf("APKBUILD is incomplete in %s", pkg.Name)
	}
	pkg.Version = pkgver + "-r" + pkgrel

	if (len(depends) != 0) {
		pkg.Dependencies = depends
	}

	return nil
}

// Scan a filename for an apk file. If one is identified, create a Package to
// represent it with all available information (Name and Version).
func find_apk(filename string) (Package, error) {
	match := pattern_apkname.FindStringSubmatch(filename)
	if (match != nil) {
		return new_package_with_version(match[1], match[2]), nil
	}
	return Package{}, fmt.Errorf("Could not identify apk in %s", filename)
}

// Construct the local directory expected to be built into.
func expected_apkdir(local_dir, arch string) string {
	if (arch == "amd64") {
		return path.Join(local_dir, "x86_64")
	} else if (arch == "arm64") {
		return path.Join(local_dir, "aarch64")
	}
	return local_dir
}

// Construct the apk filename expected to correspond to a Package.
func expected_apk(pkg Package) string {
	return fmt.Sprintf("%s-%s.apk", pkg.Name, pkg.Version)
}

// Reconstruct the relevant parts of the APKBUILD files that were parsed.
func dump_apkbuilds(packages []Package, debug_prefix string) {
	total := len(packages)
	for i, p := range packages {
		version := strings.SplitN(p.Version, "-r", 2)
		fmt.Printf("DEBUG-APK:[%d/%d] %s\n", i + 1, total, p.Name)
		fmt.Printf("DEBUG-APK:pkgver=%s\n", version[0])
		fmt.Printf("DEBUG-APK:pkgrel=%s\n", version[1])
		has_dependencies := (0 < len(p.Dependencies))
		print_if(has_dependencies, "DEBUG-APK:depends=\"")
		for _, d := range p.Dependencies {
			fmt.Printf("DEBUG-APK:\t%s\n", d)
		}
		print_if(has_dependencies, "DEBUG-APK:\"")
		fmt.Println("DEBUG-APK:")
	}
}

// Parse a string as a whitespace-delimited list.
func parse_list(list string) []string {
	return strings.Split(strings.TrimSpace(list), " ")
}

