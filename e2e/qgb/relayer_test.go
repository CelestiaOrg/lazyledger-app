package e2e

import (
	"context"
	"github.com/celestiaorg/celestia-app/x/qgb/orchestrator"
	"github.com/celestiaorg/celestia-app/x/qgb/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRelayerWithOneValidator(t *testing.T) {
	// TODO uncomment when pushing final
	//if os.Getenv("QGB_INTEGRATION_TEST") != "true" {
	//	t.Skip("Skipping QGB integration tests")
	//}
	network, err := NewQGBNetwork()
	assert.NoError(t, err)
	// preferably, run this also when ctrl+c
	defer network.DeleteAll() //nolint:errcheck
	err = network.StartMinimal()
	if err != nil {
		t.FailNow()
	}
	ctx := context.TODO()
	err = network.WaitForBlock(ctx, int64(types.DataCommitmentWindow+5))
	assert.NoError(t, err)

	err = network.WaitForOrchestratorToStart(ctx, CORE0ACCOUNTADDRESS)
	assert.NoError(t, err)

	bridge, err := network.GetLatestDeployedQGBContract(ctx)
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}

	err = network.WaitForRelayerToStart(ctx, bridge)
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}

	// FIXME should we use the evm client here or go for raw queries?
	evmClient := orchestrator.NewEvmClient(nil, *bridge, nil, network.EVMRPC)

	vsNonce, err := evmClient.StateLastValsetNonce(&bind.CallOpts{Context: ctx})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), vsNonce)

	dcNonce, err := evmClient.StateLastDataRootTupleRootNonce(&bind.CallOpts{Context: ctx})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), dcNonce)
}

func TestRelayerWithTwoValidators(t *testing.T) {
	// TODO uncomment when pushing final
	//if os.Getenv("QGB_INTEGRATION_TEST") != "true" {
	//	t.Skip("Skipping QGB integration tests")
	//}
	network, err := NewQGBNetwork()
	assert.NoError(t, err)
	// preferably, run this also when ctrl+c
	defer network.DeleteAll() //nolint:errcheck
	// start minimal network with one validator
	err = network.StartMinimal()
	if err != nil {
		t.FailNow()
	}
	// add second validator
	err = network.Start(Core1)
	if err != nil {
		t.FailNow()
	}
	// add second orchestrator
	err = network.Start(Core1Orch)
	if err != nil {
		t.FailNow()
	}

	ctx := context.TODO()
	err = network.WaitForBlock(ctx, int64(types.DataCommitmentWindow+5))
	assert.NoError(t, err)

	err = network.WaitForOrchestratorToStart(ctx, CORE0ACCOUNTADDRESS)
	assert.NoError(t, err)

	err = network.WaitForOrchestratorToStart(ctx, CORE1ACCOUNTADDRESS)
	assert.NoError(t, err)

	bridge, err := network.GetLatestDeployedQGBContract(ctx)
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}

	err = network.WaitForRelayerToStart(ctx, bridge)
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}

	// FIXME should we use the evm client here or go for raw queries?
	evmClient := orchestrator.NewEvmClient(nil, *bridge, nil, network.EVMRPC)

	dcNonce, err := evmClient.StateLastDataRootTupleRootNonce(&bind.CallOpts{Context: ctx})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), dcNonce)

	vsNonce, err := evmClient.StateLastValsetNonce(&bind.CallOpts{Context: ctx})
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), vsNonce)
}

func TestRelayerWithMultipleValidators(t *testing.T) {
	// TODO uncomment when pushing final
	//if os.Getenv("QGB_INTEGRATION_TEST") != "true" {
	//	t.Skip("Skipping QGB integration tests")
	//}
	network, err := NewQGBNetwork()
	assert.NoError(t, err)
	// preferably, run this also when ctrl+c
	defer network.DeleteAll() //nolint:errcheck
	// start full network with four validatorS
	err = network.StartAll()
	if err != nil {
		t.FailNow()
	}

	ctx := context.TODO()
	err = network.WaitForBlock(ctx, int64(types.DataCommitmentWindow+5))
	assert.NoError(t, err)

	// check whether the four validators are up and running
	querier, err := orchestrator.NewQuerier(network.CelestiaGRPC, network.TendermintRPC, nil)
	assert.NoError(t, err)

	err = network.WaitForOrchestratorToStart(ctx, CORE0ACCOUNTADDRESS)
	assert.NoError(t, err)

	err = network.WaitForOrchestratorToStart(ctx, CORE1ACCOUNTADDRESS)
	assert.NoError(t, err)

	err = network.WaitForOrchestratorToStart(ctx, CORE2ACCOUNTADDRESS)
	assert.NoError(t, err)

	err = network.WaitForOrchestratorToStart(ctx, CORE3ACCOUNTADDRESS)
	assert.NoError(t, err)

	lastValsets, err := querier.QueryLastValsets(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(lastValsets[0].Members))

	bridge, err := network.GetLatestDeployedQGBContract(ctx)
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}

	err = network.WaitForRelayerToStart(ctx, bridge)
	if err != nil {
		assert.NoError(t, err)
		t.FailNow()
	}

	// FIXME should we use the evm client here or go for raw queries?
	evmClient := orchestrator.NewEvmClient(nil, *bridge, nil, network.EVMRPC)

	dcNonce, err := evmClient.StateLastDataRootTupleRootNonce(&bind.CallOpts{Context: ctx})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), dcNonce)

	vsNonce, err := evmClient.StateLastValsetNonce(&bind.CallOpts{Context: ctx})
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, uint64(2), vsNonce)
}
