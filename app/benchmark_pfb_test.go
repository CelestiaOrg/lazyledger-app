package app_test

import (
	"fmt"
	"github.com/celestiaorg/celestia-app/v3/app"
	"github.com/celestiaorg/celestia-app/v3/app/encoding"
	"github.com/celestiaorg/celestia-app/v3/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/v3/pkg/user"
	testutil "github.com/celestiaorg/celestia-app/v3/test/util"
	"github.com/celestiaorg/celestia-app/v3/test/util/testfactory"
	"github.com/celestiaorg/go-square/v2/share"
	blobtx "github.com/celestiaorg/go-square/v2/tx"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/proto/tendermint/version"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkCheckTx_PFB_Multi(b *testing.B) {
	testCases := []struct {
		size int
	}{
		{size: 300},
		{size: 500},
		{size: 1000},
		{size: 5000},
		{size: 10_000},
		{size: 50_000},
		{size: 100_000},
		{size: 200_000},
		{size: 300_000},
		{size: 400_000},
		{size: 500_000},
		{size: 1_000_000},
		{size: 2_000_000},
		{size: 3_000_000},
		{size: 4_000_000},
		{size: 5_000_000},
		{size: 6_000_000},
	}
	for _, testCase := range testCases {
		b.Run(fmt.Sprintf("%d bytes", testCase.size), func(b *testing.B) {
			benchmarkCheckTx_PFB(b, testCase.size)
		})
	}
}

func benchmarkCheckTx_PFB(b *testing.B, size int) {
	testApp, rawTxs := generatePayForBlobTransactions(b, 1, size)
	testApp.Commit()

	checkTxRequest := types.RequestCheckTx{
		Tx:   rawTxs[0],
		Type: types.CheckTxType_New,
	}

	b.ResetTimer()
	resp := testApp.CheckTx(checkTxRequest)
	b.StopTimer()
	require.Equal(b, uint32(0), resp.Code)
	require.Equal(b, "", resp.Codespace)
	b.ReportMetric(float64(resp.GasUsed), "gas_used")
	b.ReportMetric(float64(len(rawTxs[0])), "transaction_size(byte)")
}

func BenchmarkDeliverTx_PFB_Multi(b *testing.B) {
	testCases := []struct {
		size int
	}{
		{size: 300},
		{size: 500},
		{size: 1000},
		{size: 5000},
		{size: 10_000},
		{size: 50_000},
		{size: 100_000},
		{size: 200_000},
		{size: 300_000},
		{size: 400_000},
		{size: 500_000},
		{size: 1_000_000},
		{size: 2_000_000},
		{size: 3_000_000},
		{size: 4_000_000},
		{size: 5_000_000},
		{size: 6_000_000},
	}
	for _, testCase := range testCases {
		b.Run(fmt.Sprintf("%d bytes", testCase.size), func(b *testing.B) {
			benchmarkDeliverTx_PFB(b, testCase.size)
		})
	}
}

func benchmarkDeliverTx_PFB(b *testing.B, size int) {
	testApp, rawTxs := generatePayForBlobTransactions(b, 1, size)

	blobTx, ok, err := blobtx.UnmarshalBlobTx(rawTxs[0])
	require.NoError(b, err)
	require.True(b, ok)

	deliverTxRequest := types.RequestDeliverTx{
		Tx: blobTx.Tx,
	}

	b.ResetTimer()
	resp := testApp.DeliverTx(deliverTxRequest)
	b.StopTimer()
	require.Equal(b, uint32(0), resp.Code)
	require.Equal(b, "", resp.Codespace)
	b.ReportMetric(float64(resp.GasUsed), "gas_used")
	b.ReportMetric(float64(len(rawTxs[0])), "transaction_size(byte)")
}

func BenchmarkPrepareProposal_PFB_Multi(b *testing.B) {
	testCases := []struct {
		count, size int
	}{
		{count: 15_000, size: 300},
		{count: 10_000, size: 500},
		{count: 6_000, size: 1000},
		{count: 3_000, size: 5000},
		{count: 1_000, size: 10_000},
		{count: 500, size: 50_000},
		{count: 100, size: 100_000},
		{count: 100, size: 200_000},
		{count: 50, size: 300_000},
		{count: 50, size: 400_000},
		{count: 30, size: 500_000},
		{count: 10, size: 1_000_000},
		{count: 5, size: 2_000_000},
		{count: 3, size: 3_000_000},
		{count: 3, size: 4_000_000},
		{count: 2, size: 5_000_000},
		{count: 2, size: 6_000_000},
	}
	for _, testCase := range testCases {
		b.Run(fmt.Sprintf("%d transactions of %d bytes", testCase.count, testCase.size), func(b *testing.B) {
			benchmarkPrepareProposal_PFB(b, testCase.count, testCase.size)
		})
	}
}

func benchmarkPrepareProposal_PFB(b *testing.B, count, size int) {
	testApp, rawTxs := generatePayForBlobTransactions(b, count, size)

	blockData := &tmproto.Data{
		Txs: rawTxs,
	}
	prepareProposalRequest := types.RequestPrepareProposal{
		BlockData: blockData,
		ChainId:   testApp.GetChainID(),
		Height:    10,
	}

	b.ResetTimer()
	prepareProposalResponse := testApp.PrepareProposal(prepareProposalRequest)
	b.StopTimer()
	require.GreaterOrEqual(b, len(prepareProposalResponse.BlockData.Txs), 1)
	b.ReportMetric(float64(b.Elapsed().Nanoseconds()), "prepare_proposal_time(ns)")
	b.ReportMetric(float64(len(prepareProposalResponse.BlockData.Txs)), "number_of_transactions")
	b.ReportMetric(float64(len(rawTxs[0])), "transactions_size(byte)")
	b.ReportMetric(calculateBlockSizeInMb(prepareProposalResponse.BlockData.Txs), "block_size(mb)")
	b.ReportMetric(float64(calculateTotalGasUsed(testApp, rawTxs)), "total_gas_used")
}

func BenchmarkProcessProposal_PFB_Multi(b *testing.B) {
	testCases := []struct {
		count, size int
	}{
		{count: 15_000, size: 300},
		{count: 10_000, size: 500},
		{count: 6_000, size: 1000},
		{count: 3_000, size: 5000},
		{count: 1_000, size: 10_000},
		{count: 500, size: 50_000},
		{count: 100, size: 100_000},
		{count: 100, size: 200_000},
		{count: 50, size: 300_000},
		{count: 50, size: 400_000},
		{count: 30, size: 500_000},
		{count: 10, size: 1_000_000},
		{count: 5, size: 2_000_000},
		{count: 3, size: 3_000_000},
		{count: 3, size: 4_000_000},
		{count: 2, size: 5_000_000},
		{count: 2, size: 6_000_000},
	}
	for _, testCase := range testCases {
		b.Run(fmt.Sprintf("%d transactions of %d bytes", testCase.count, testCase.size), func(b *testing.B) {
			benchmarkProcessProposal_PFB(b, testCase.count, testCase.size)
		})
	}
}

func benchmarkProcessProposal_PFB(b *testing.B, count, size int) {
	testApp, rawTxs := generatePayForBlobTransactions(b, count, size)

	blockData := &tmproto.Data{
		Txs: rawTxs,
	}
	prepareProposalRequest := types.RequestPrepareProposal{
		BlockData: blockData,
		ChainId:   testApp.GetChainID(),
		Height:    10,
	}

	prepareProposalResponse := testApp.PrepareProposal(prepareProposalRequest)
	require.GreaterOrEqual(b, len(prepareProposalResponse.BlockData.Txs), 1)

	processProposalRequest := types.RequestProcessProposal{
		BlockData: prepareProposalResponse.BlockData,
		Header: tmproto.Header{
			Height:   10,
			DataHash: prepareProposalResponse.BlockData.Hash,
			ChainID:  testutil.ChainID,
			Version: version.Consensus{
				App: testApp.AppVersion(),
			},
		},
	}

	b.ResetTimer()
	resp := testApp.ProcessProposal(processProposalRequest)
	b.StopTimer()
	require.Equal(b, types.ResponseProcessProposal_ACCEPT, resp.Result)

	b.ReportMetric(float64(b.Elapsed().Nanoseconds()), "process_proposal_time(ns)")
	b.ReportMetric(float64(len(prepareProposalResponse.BlockData.Txs)), "number_of_transactions")
	b.ReportMetric(float64(len(rawTxs[0])), "transactions_size(byte)")
	b.ReportMetric(calculateBlockSizeInMb(prepareProposalResponse.BlockData.Txs), "block_size(mb)")
	b.ReportMetric(float64(calculateTotalGasUsed(testApp, rawTxs)), "total_gas_used")
}

func BenchmarkProcessProposal_PFB_Half_Second(b *testing.B) {
	testCases := []struct {
		count, size int
	}{
		{count: 11_000, size: 50},
		{count: 11_000, size: 100},
		{count: 11_000, size: 200},
		{count: 11_000, size: 300},
		{count: 11_000, size: 400},
		{count: 7000, size: 500},
		{count: 7000, size: 600},
		{count: 5000, size: 1_000},
		{count: 5000, size: 1200},
		{count: 5000, size: 1500},
		{count: 5000, size: 1800},
		{count: 5000, size: 2000},
	}
	for _, testCase := range testCases {
		b.Run(fmt.Sprintf("%d transactions of %d bytes", testCase.count, testCase.size), func(b *testing.B) {
			benchmarkProcessProposal_PFB_Half_Second(b, testCase.count, testCase.size)
		})
	}
}

func benchmarkProcessProposal_PFB_Half_Second(b *testing.B, count, size int) {
	testApp, rawTxs := generatePayForBlobTransactions(b, count, size)

	targetTimeLowerBound := 0.499
	targetTimeUpperBound := 0.511

	start := 0
	end := count
	segment := end - start
	for {
		if segment == 1 {
			break
		}

		prepareProposalRequest := types.RequestPrepareProposal{
			BlockData: &tmproto.Data{
				Txs: rawTxs[start:end],
			},
			ChainId: testApp.GetChainID(),
			Height:  10,
		}
		prepareProposalResponse := testApp.PrepareProposal(prepareProposalRequest)
		require.GreaterOrEqual(b, len(prepareProposalResponse.BlockData.Txs), 1)

		processProposalRequest := types.RequestProcessProposal{
			BlockData: prepareProposalResponse.BlockData,
			Header: tmproto.Header{
				Height:   10,
				DataHash: prepareProposalResponse.BlockData.Hash,
				ChainID:  testutil.ChainID,
				Version: version.Consensus{
					App: testApp.AppVersion(),
				},
			},
		}

		startTime := time.Now()
		resp := testApp.ProcessProposal(processProposalRequest)
		endTime := time.Now()
		require.Equal(b, types.ResponseProcessProposal_ACCEPT, resp.Result)

		timeElapsed := float64(endTime.Sub(startTime).Nanoseconds()) / 1e9

		b.ReportMetric(timeElapsed, fmt.Sprintf("processProposalTime(s)_%d_%d_%f", end-start, size, calculateBlockSizeInMb(prepareProposalResponse.BlockData.Txs[start:end])))

		if timeElapsed < targetTimeLowerBound {
			end = end + segment/2
			segment = end - start
		} else if timeElapsed > targetTimeUpperBound {
			end = end / 2
			segment = end - start
		} else {
			break
		}
	}
}

// generatePayForBlobTransactions creates a test app then generates a number
// of valid PFB transactions.
func generatePayForBlobTransactions(b *testing.B, count int, size int) (*app.App, [][]byte) {
	account := "test"
	testApp, kr := testutil.SetupTestAppWithGenesisValSet(app.DefaultConsensusParams(), account)
	addr := testfactory.GetAddress(kr, account)
	enc := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	acc := testutil.DirectQueryAccount(testApp, addr)
	accountSequence := acc.GetSequence()
	signer, err := user.NewSigner(kr, enc.TxConfig, testutil.ChainID, appconsts.LatestVersion, user.NewAccount(account, acc.GetAccountNumber(), acc.GetSequence()))
	require.NoError(b, err)

	rawTxs := make([][]byte, 0, count)
	randomBytes := make([]byte, size)
	_, err = rand.Read(randomBytes)
	require.NoError(b, err)
	blob, err := share.NewBlob(share.RandomNamespace(), randomBytes, 1, acc.GetAddress().Bytes())
	require.NoError(b, err)
	for i := 0; i < count; i++ {
		tx, _, err := signer.CreatePayForBlobs(account, []*share.Blob{blob}, user.SetGasLimit(2549760000), user.SetFee(10000))
		require.NoError(b, err)
		rawTxs = append(rawTxs, tx)
		accountSequence++
		err = signer.SetSequence(account, accountSequence)
		require.NoError(b, err)
	}
	return testApp, rawTxs
}
