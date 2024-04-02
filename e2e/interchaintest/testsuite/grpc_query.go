package testsuite

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

var queryReqToPath = make(map[string]string)

func populateQueryReqToPath(ctx context.Context, chain *cosmos.CosmosChain) error {
	resp, err := queryFileDescriptors(ctx, chain)
	if err != nil {
		return err
	}

	for _, fileDescriptor := range resp.Files {
		for _, service := range fileDescriptor.GetService() {
			fmt.Println("Service Uninterpreted Options: ", service.GetOptions().GetUninterpretedOption())
			for _, method := range service.GetMethod() {
				queryReqToPath[method.GetInputType()] = service.GetName() + "/" + method.GetName()
			}
		}
	}

	fmt.Println("queryReqToPath: ", queryReqToPath)

	return nil
}

// Queries the chain with a query request and deserializes the response to T
func GRPCQuery[T any](ctx context.Context, chain *cosmos.CosmosChain, req proto.Message, opts ...grpc.CallOption) (*T, error) {
	path, ok := queryReqToPath[proto.MessageName(req)]
	if !ok {
		return nil, fmt.Errorf("no path found for %s", proto.MessageName(req))
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

func queryFileDescriptors(ctx context.Context, chain *cosmos.CosmosChain) (*reflectionv1.FileDescriptorsResponse, error) {
	// Create a connection to the gRPC server.
	grpcConn, err := grpc.Dial(
		chain.GetHostGRPCAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	defer grpcConn.Close()

	resp := new(reflectionv1.FileDescriptorsResponse)
	err = grpcConn.Invoke(
		ctx, reflectionv1.ReflectionService_FileDescriptors_FullMethodName,
		&reflectionv1.FileDescriptorsRequest{}, resp,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
