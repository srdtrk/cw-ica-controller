package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/srdtrk/go-codegen/e2esuite/v8/e2esuite"
	"github.com/srdtrk/go-codegen/e2esuite/v8/types/cwicacontroller"
)

// ContactTestSuite is a suite of tests that wraps the TestSuite
// and can provide additional functionality
type ContractTestSuite struct {
	e2esuite.TestSuite

	// this line is used by go-codegen # suite/contract
}

// SetupSuite calls the underlying ContractTestSuite's SetupSuite method
func (s *ContractTestSuite) SetupSuite(ctx context.Context) {
	s.TestSuite.SetupSuite(ctx)
}

// TestWithContractTestSuite is the boilerplate code that allows the test suite to be run
func TestWithContractTestSuite(t *testing.T) {
	suite.Run(t, new(ContractTestSuite))
}

// TestContract is an example test function that will be run by the test suite
func (s *ContractTestSuite) TestContract() {
	ctx := context.Background()

	s.SetupSuite(ctx)

	wasmd1 := s.ChainA

	// Add your test code here. For example, upload and instantiate a contract:
	// This boilerplate may be moved to SetupSuite if it is common to all tests.
	var contract *cwicacontroller.Contract
	s.Run("UploadAndInstantiateContract", func() {
		// Upload the contract code to the chain.
		codeID, err := wasmd1.StoreContract(ctx, s.UserA.KeyName(), "./relative/path/to/your_contract.wasm", "--gas", "500000")
		s.Require().NoError(err)

		// Instantiate the contract using contract helpers.
		// This will an error if the instantiate message is invalid.
		contract, err = cwicacontroller.Instantiate(ctx, s.UserA.KeyName(), codeID, "", wasmd1, cwicacontroller.InstantiateMsg{})
		s.Require().NoError(err)

		s.Require().NotEmpty(contract.Address)
	})
}
