# Simple Builder

A simple package builder.
It leverages `docker(1)` for builds.
It leverages `rsync(1)` for file transfers.

```
simple-builder -repository /local/path/to/package/repository/alpine/v3.17/aarch64
```

It autodetects architecture based on the repository, although an option for
explicit architecture is also available.

```
simple-builder -repository user:/var/pkgs -architecture amd64
```

**Note:
Only supports ARM64 (a.k.a. AArch64 or ARM64v8) and AMD64 (a.k.a. x86_64)
currently.**

It scans a package source directory for packages that could be built.
This defaults to `./src` but can be configured 

```
simple-builder -repository rsync://user@example.com:8888 -architecture arm64 -source /usr/local/src
```

**Note:
The package source directory should be structured like `$src/$package/*`.**

It parses package source files (i.e. `APKBUILD`s, etc.) for versioned packages
that should exist.
If a versioned package does not exist, it queues those package to be built.

It also understands dependencies.
It will return an error if there is a circular dependency.
If a package has been built, it asserts that any other packages which depend
on the first one should be updated as well.
It will similarly return an error if a breaking build might be queued.

Packages are built into a local folder.
This defaults to `./pkg` and but can be configured.

```
simple-builder -repository host:/var/alpine/v3.17/x86_64 -destination /var/alpine/v3.17/x86_64
```

On success, both the package and APKINDEX files are pushed to the repository
immediately.

It offers a simple command line interface.
Calling the binary without a command option will cause the program to print
summary information and exit.
Calling with the `-build` option will begin running through the build queue.
Use the `-verbose` flag to get diagnostic information.
Try `-help` for more information about all of this.


## License

I license this work under the BSD 3-clause license.

