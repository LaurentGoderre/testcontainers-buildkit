package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/docker/docker/api/types/container"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestBuildKit(t *testing.T) {
	ctx := context.Background()

	context, _ := filepath.Abs(".")

	tmpDir, err := os.MkdirTemp("", "buildkit-out")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tag := "test"

	req := tc.ContainerRequest{
		Image:      "moby/buildkit:buildx-stable-1",
		Entrypoint: []string{"buildctl-daemonless.sh"},
		Cmd: []string{
			"build",
			"--frontend", "dockerfile.v0",
			"--local", "context=/work",
			"--local", "dockerfile=/work",
			"--output", fmt.Sprintf("type=docker,name=%s,dest=/out/out.tar", tag),
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Privileged = true
			hc.Binds = append(hc.Binds,
				context+":/work",
				tmpDir+":/out",
			)
		},
		WaitingFor: wait.ForAll(wait.ForExit(), wait.ForLog("DONE")),
	}
	buildkit, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	tc.CleanupContainer(t, buildkit)
	require.NoError(t, err)

	provider, err := tc.ProviderDocker.GetProvider()
	require.NoError(t, err)

	file, err := os.Open(filepath.Join(tmpDir, "out.tar"))
	require.NoError(t, err)

	docker, _ := provider.(*tc.DockerProvider)
	_, err = docker.Client().ImageLoad(ctx, file)
	require.NoError(t, err)
}
