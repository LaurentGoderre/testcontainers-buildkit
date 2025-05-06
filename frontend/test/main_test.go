package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/docker/docker/api/types/container"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestBuildKit(t *testing.T) {
	ctx := context.Background()

	self, _ := filepath.Abs(".")
	frontend := filepath.Join(self, "..", "src")

	frontendOut, err := os.MkdirTemp("", "buildkit-out")
	require.NoError(t, err)
	defer os.RemoveAll(frontendOut)

	testOut, err := os.MkdirTemp(filepath.Join(self, ".."), "out-")
	require.NoError(t, err)

	req := tc.ContainerRequest{
		Image: "moby/buildkit:buildx-stable-1",
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Privileged = true
			hc.Binds = append(hc.Binds,
				self+":/test",
				frontend+":/frontend",
				frontendOut+":/tmp/frontend",
				testOut+":/out",
			)
		},
		WaitingFor: wait.ForLog("running server on"),
	}
	buildkit, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	tc.CleanupContainer(t, buildkit)
	require.NoError(t, err)

	frontendBuildCmd := []string{
		"buildctl",
		"build",
		"--frontend", "dockerfile.v0",
		"--local", "context=/frontend",
		"--local", "dockerfile=/frontend",
		"--output", "type=oci,tar=false,dest=/tmp/frontend",
	}

	exitCode, _, err := buildkit.Exec(ctx, frontendBuildCmd)
	require.NoError(t, err)
	require.Equal(t, 0, exitCode)

	indexFile, err := os.Open(filepath.Join(frontendOut, "index.json"))
	require.NoError(t, err)
	byteValue, _ := io.ReadAll(indexFile)

	var index ocispecs.Index

	json.Unmarshal(byteValue, &index)
	require.Equal(t, 1, len(index.Manifests))

	digest := index.Manifests[0].Digest

	testBuildCommand := []string{
		"buildctl",
		"build",
		"--frontend", "dockerfile.v0",
		"--opt", "context:frontend=oci-layout://frontend@" + digest.String(),
		"--opt", "build-arg:BUILDKIT_SYNTAX=frontend",
		"--local", "context=/test",
		"--local", "dockerfile=/test",
		"--oci-layout", "frontend=/tmp/frontend",
		"--output", "type=local,dest=/out",
	}

	exitCode, _, err = buildkit.Exec(ctx, testBuildCommand)
	require.NoError(t, err)
	require.Equal(t, 0, exitCode)
}
