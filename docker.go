package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// Create a container for building a package, start the build, and branch
// based on the result.
func build_package(pkg Package, srcdir, pkgdir, arch string) error {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if (err != nil) {
		return err
	}

	conf := container.Config{
		Image: "registry.intra.dominic-ricottone.com/apkbuilder:latest",
		Cmd: []string{pkg.Name},
	}

	con_conf := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type: mount.TypeBind,
				Source: srcdir,
				Target: "/home/builder/src",
			},
			{
				Type: mount.TypeBind,
				Source: pkgdir,
				Target: "/home/builder/packages/src",
			},
		},
	}

	plats := specs.Platform{
		Architecture: arch,
		OS: "linux",
	}

	con, err := cli.ContainerCreate(ctx, &conf, &con_conf, nil, &plats, "")
	if (err != nil) {
		return err
	}

	start_opts := types.ContainerStartOptions{}

	cli.ContainerStart(ctx, con.ID, start_opts)

	err = check_result(cli, ctx, con.ID)
	if (err != nil) {
		return err
	}

	rm_opts := types.ContainerRemoveOptions{
		Force: true,
	}

	cli.ContainerRemove(ctx, con.ID, rm_opts)

	return nil
}

// Get the result of a build. Blocks until the build is complete.
func check_result(cli *client.Client, ctx context.Context, id string) error {
	statusC, errC := cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	sigC := make(chan os.Signal)
	signal.Notify(sigC, os.Interrupt)

	select {
	case _ = <-sigC:
		return errors.New("Build interrupted")

	case err := <-errC:
		if (err != nil) {
			return err
		}

	case status := <-statusC:
		if status.StatusCode != 0 {
			dump_logs(cli, ctx, id)
			return errors.New("Build failed")
		}
	}

	return nil
}

// Dump logs from a build.
func dump_logs(cli *client.Client, ctx context.Context, id string) {
	conf := types.ContainerLogsOptions{
		ShowStdout: true,
	}

	out, err := cli.ContainerLogs(ctx, id, conf)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

