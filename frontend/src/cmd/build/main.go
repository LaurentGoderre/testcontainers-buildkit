package main

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/frontend/gateway/grpcclient"
	"github.com/moby/buildkit/util/appcontext"
)

func main() {
	if err := grpcclient.RunFromEnvironment(appcontext.Context(), build); err != nil {
		fmt.Errorf("fatal error: %+v", err)
		panic(err)
	}
}

func build(ctx context.Context, c client.Client) (*client.Result, error) {
	root := llb.Scratch().File(
		llb.Mkfile("/foo", 0644, []byte("hello")),
	)

	def, err := root.Marshal(ctx)
	if err != nil {
		return nil, err
	}

	return c.Solve(ctx, client.SolveRequest{
		Definition: def.ToPB(),
	})
}
