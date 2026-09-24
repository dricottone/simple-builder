package main

import (
	"encoding/binary"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
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

	create_opts := client.ContainerCreateOptions{
		Config: &container.Config{
			Image: "registry.intra.dominic-ricottone.com/apkbuilder:latest",
			Cmd: []string{pkg.Name},
		},
		HostConfig: &container.HostConfig{
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
		},
		Platform: &specs.Platform{
			Architecture: arch,
			OS: "linux",
		},
		Name: "",
	}

	con, err := cli.ContainerCreate(ctx, create_opts)
	if (err != nil) {
		return err
	}

	cli.ContainerStart(ctx, con.ID, client.ContainerStartOptions{})
	err = check_result(cli, ctx, con.ID)

	rm_opts := client.ContainerRemoveOptions{
		Force: true,
	}

	cli.ContainerRemove(ctx, con.ID, rm_opts)
	if (err != nil) {
		return err
	}
	return nil
}

// Get the result of a build. Blocks until the build is complete.
func check_result(cli *client.Client, ctx context.Context, id string) error {
	opts := client.ContainerWaitOptions{
		Condition: container.WaitConditionNotRunning,
	}
	wait := cli.ContainerWait(ctx, id, opts)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	select {
	case _ = <-sig:
		return errors.New("Build interrupted")

	case err := <-wait.Error:
		if (err != nil) {
			return err
		}

	case status := <-wait.Result:
		if status.StatusCode != 0 {
			dump_logs(cli, ctx, id)
			return errors.New("Build failed")
		}
	}

	return nil
}

// Dump logs from a build.
func dump_logs(cli *client.Client, ctx context.Context, id string) {
	opts := client.ContainerLogsOptions{
		ShowStdout: true,
	}
	out, err := cli.ContainerLogs(ctx, id, opts)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Docker log lines have an 8-byte header
	//  + First byte is stream identifier (e.g., 1=STDOUT)
	//  + Next three bytes unused (for now)
	//  + Final four bytes are length of message
	header := make([]byte, 8)
	for {
		_, err := out.Read(header)
		if err != nil {
			panic(err)
		}

		length := binary.BigEndian.Uint32(header[4:])

		msg := make([]byte, length)
		_, err = out.Read(msg)
		fmt.Printf(string(msg))
	}
}

