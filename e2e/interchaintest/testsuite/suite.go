package testsuite

import (
	"context"

	interchaintest "github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"

	dockerclient "github.com/docker/docker/client"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

type ContractTestSuite struct {
	suite.Suite

	ChainA *cosmos.CosmosChain
	ChainB *cosmos.CosmosChain
	UserA ibc.Wallet
	UserB ibc.Wallet
	dockerClient   *dockerclient.Client
	Relayer ibc.Relayer
	network string
	logger *zap.Logger
}

func (s *ContractTestSuite) SetupSuite(chainSpecs []*interchaintest.ChainSpec) {
	if len(chainSpecs) != 2 {
		panic("ContractTestSuite requires exactly 2 chain specs")
	}

	t := s.T()
	ctx := context.Background()

	s.logger = zaptest.NewLogger(t)
	s.dockerClient, s.network = interchaintest.DockerSetup(t)

	cf := interchaintest.NewBuiltinChainFactory(s.logger, chainSpecs)

	chains, err := cf.Chains(t.Name())
	s.Require().NoError(err)
	s.ChainA, s.ChainB = chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// This is currently the only relayer image that works with the main (next) version of ibc-go 
	customRelayerImage := relayer.CustomDockerImage("damiannolan/rly", "", "100:1000")

	// Fund a user account on ChainA and ChainB
	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, s.ChainA, s.ChainB)
	s.UserA = users[0]
	s.UserB = users[1]

	s.Relayer = interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayer.RelayerOptionExtraStartFlags{Flags: []string{"-p", "events", "-b", "100"}},
		customRelayerImage,
	).Build(t, s.dockerClient, s.network)
}

func (s *ContractTestSuite) SetupClientsAndConnections() {
	t := s.T()
	ctx := context.Background()
	eRep := testreporter.NewNopReporter().RelayerExecReporter(t)

	pathName := s.ChainA.Config().Name + "-" + s.ChainB.Config().Name

	ic := interchaintest.NewInterchain().
		AddChain(s.ChainA).
		AddChain(s.ChainB).
		AddRelayer(s.Relayer, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  s.ChainA,
			Chain2:  s.ChainB,
			Relayer: s.Relayer,
			Path:    pathName,
		})

	s.Require().NoError(ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           s.dockerClient,
		NetworkID:        s.network,
		// I don't exactly know the consequences of SkipPathCreation
		SkipPathCreation: true,
	}))

	// Generate a new IBC path
	err := s.Relayer.GeneratePath(ctx, eRep, s.ChainA.Config().ChainID, s.ChainB.Config().ChainID, pathName)
	s.Require().NoError(err)

	// Create new clients
	err = s.Relayer.CreateClients(ctx, eRep, pathName, ibc.CreateClientOptions{TrustingPeriod: "330h"})
	s.Require().NoError(err)

	err = testutil.WaitForBlocks(ctx, 2, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// Create a new connection
	err = s.Relayer.CreateConnections(ctx, eRep, pathName)
	s.Require().NoError(err)

	err = testutil.WaitForBlocks(ctx, 2, s.ChainA, s.ChainB)
	s.Require().NoError(err)

	// Query for the newly created connection in wasmd
	connections, err := s.Relayer.GetConnections(ctx, eRep, s.ChainA.Config().ChainID)
	s.Require().NoError(err)
	require.Equal(t, 1, len(connections))
	wasmdConnection := connections[0]
	require.Equal(t, "connection-0", wasmdConnection.ID)

	// Query for the newly created connection in simd
	connections, err = s.Relayer.GetConnections(ctx, eRep, s.ChainB.Config().ChainID)
	s.Require().NoError(err)
	// localhost is always a connection in main (next) version of ibc-go
	require.Equal(t, 2, len(connections))
	simdConnection := connections[0]
	require.Equal(t, "connection-0", simdConnection.ID)
}