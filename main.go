package main

import (
	"flag"
	"fmt"
)

var (
	verbose = flag.Bool("verbose", false, "Show debugging messages")
	build = flag.Bool("build", false, "Build packages")
	summary = flag.Bool("summary", false, "Summarize packages to build")
	source = flag.String("source", "./src", "Directory of package sources")
	destination = flag.String("destination", "./pkg", "Directory of packages")
	architecture = flag.String("architecture", "detected from repository", "architecture to build")
	repository = flag.String("repository", "", "Connection string for the remote package repository")
)

// Conditionally print a string.
func print_if(condition bool, str string) {
	if (condition == true) {
		fmt.Println(str)
	}
}

// Print debugging information if -verbose was passed to the program.
func debug(str string) {
	print_if(*verbose, str)
}

// Identify Packages in the package source directory.
func list_package_sources(local_dir string) ([]Package, error) {
	packages, err := walk_package_sources(local_dir)
	if (err != nil) {
		return nil, err
	}

	if (*verbose == true) {
		dump_apkbuilds(packages, "DEBUG-MAIN")
	}

	return packages, nil
}

// Identify Packages in the repository.
func list_repository(remote_dir string) ([]Package, error) {
	packages, err := fetch_repository_listing(remote_dir)
	if (err != nil) {
		return nil, err
	}

	if (*verbose == true) {
		dump_apkbuilds(packages, "DEBUG-MAIN")
	}

	return packages, nil
}

// Compare Packages between the package source directory and the repository.
func compare_lists(local_dir, remote_dir string) ([]Package, error) {
	queue := []Package{}
	had_errors := false

	package_sources, err := list_package_sources(local_dir)
	if (err != nil) {
		return nil, err
	}

	repository, err := list_repository(remote_dir)
	if (err != nil) {
		return nil, err
	}

	for i, _ := range package_sources {
		err = find_builds(&package_sources[i], &repository)
		if (err != nil) {
			return nil, err
		}

		if (package_sources[i].Build == true) {
			queue = append(queue, package_sources[i])
		}

		if (package_sources[i].Error == true) {
			print_if(!had_errors, "Warnings:")
			had_errors = true
			fmt.Printf("%s %s - %s\n", package_sources[i].Name, package_sources[i].Version, package_sources[i].Message)
		}
	}

	err = find_breaking_builds(&queue)
	if (err != nil) {
		return nil, err
	}

	err = sort_queue(&queue)
	if (err != nil) {
		return nil, err
	}

	return queue, nil
}

// Sort the Package list in-place.
func sort_queue(packages *[]Package) error {
	visited := []string{}
	resolved := []Package{}

	for i, _ := range (*packages) {
		err := resolve_dependencies(&resolved, packages, i, &visited)
		if (err != nil) {
			return err
		}
	}

	(*packages) = resolved
	return nil
}

// Build Packages.
func build_packages(packages []Package, source, destination, arch, repository string) error {
	for _, pkg := range packages {
		debug(fmt.Sprintf("Building %s...", pkg.Name))
		err := build_package(pkg, source, destination, arch)
		if (err != nil) {
			return err
		}

		debug(fmt.Sprintf("Pushing %s...", pkg.Name))
		local_dir := expected_apkdir(destination, arch)
		err = push_package(pkg, local_dir, repository)
		if (err != nil) {
			return err
		}
	}

	return nil
}

// Print details about Packages queued for build.
func summarize_packages(packages []Package) {
	if (len(packages) == 0) {
		fmt.Println("Nothing to do")
	} else {
		fmt.Println("Packages to build:")
		for _, p := range packages {
			fmt.Printf("  %s %s - %s\n", p.Name, p.Version, p.Message)
		}
		fmt.Println("To start building, pass the `-build` option")
	}
}

func main() {
	flag.Parse()
	src := clean_source(*source)
	pkg := clean_destination(*destination)
	repo := clean_repository(*repository)
	arch := clean_architecture(*architecture, repo)

	packages, err := compare_lists(src, repo)
	if (err != nil) {
		panic(err)
	}

	if (*build == true) {
		err = build_packages(packages, src, pkg, arch, repo)
		if (err != nil) {
			panic(err)
		}
	} else {
		summarize_packages(packages)
	}
}

