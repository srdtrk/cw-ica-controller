package testsuite

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// Queries the chain with a query request and deserializes the response to T
func GRPCQuery[T any](ctx context.Context, chain *cosmos.CosmosChain, req proto.Message, opts ...grpc.CallOption) (*T, error) {
	path, err := getProtoPath(req)
	if err != nil {
		return nil, err
	}

	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	defer grpcConn.Close()

	resp := new(T)
	err = grpcConn.Invoke(ctx, path, req, resp, opts...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func getProtoPath(req proto.Message) (string, error) {
	typeUrl := "/" + proto.MessageName(req)

	queryIndex := strings.Index(typeUrl, "Query")
	if queryIndex == -1 {
		return "", fmt.Errorf("invalid typeUrl: %s", typeUrl)
	}

	// Add to the index to account for the length of "Query"
	queryIndex += len("Query")

	// Add a slash before the query
	urlWithSlash := typeUrl[:queryIndex] + "/" + typeUrl[queryIndex:]
	if !strings.HasSuffix(urlWithSlash, "Request") {
		return "", fmt.Errorf("invalid typeUrl: %s", typeUrl)
	}

	return strings.TrimSuffix(urlWithSlash, "Request"), nil
}
