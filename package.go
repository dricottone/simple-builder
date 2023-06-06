package main

// Package stores the core information about a software package.
type Package struct {
	Name         string
	Version      string
	Dependencies []string
	Message      string
	Build        bool
	Error        bool
}

func new_package(name string) Package {
	return Package{name, "", []string{}, "", false, false}
}

func new_package_with_version(name, version string) Package {
	return Package{name, version, []string{}, "", false, false}
}

