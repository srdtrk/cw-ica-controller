package testsuite

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
)

type GRPCQuerier[T any] struct {
	t     *testing.T
	chain *cosmos.CosmosChain
	path  string
}

func NewGRPCQuerier[T any](t *testing.T, chain *cosmos.CosmosChain, path string) *GRPCQuerier[T] {
	t.Helper()
	return &GRPCQuerier[T]{
		t:     t,
		chain: chain,
		path:  path,
	}
}

func (q *GRPCQuerier[T]) GRPCQuery(ctx context.Context, req interface{}, opts ...grpc.CallOption) (*T, error) {
	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		q.chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(q.t, err)

	q.t.Cleanup(func() {
		if err := grpcConn.Close(); err != nil {
			q.t.Logf("failed closing GRPC connection to chain %s: %s", q.chain.Config().ChainID, err)
		}
	})

	resp := new(T)
	err = grpcConn.Invoke(ctx, q.path, req, resp, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
