package app_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	core "github.com/tendermint/tendermint/proto/tendermint/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	coretypes "github.com/tendermint/tendermint/types"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/pkg/da"
	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/celestiaorg/celestia-app/testutil"
	"github.com/celestiaorg/celestia-app/testutil/blobfactory"
)

func TestBlobInclusionCheck(t *testing.T) {
	testApp, _ := testutil.SetupTestAppWithGenesisValSet()
	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)

	// block with all blobs included
	validData := func() *core.Data {
		return &core.Data{
			Txs: coretypes.Txs(blobfactory.RandBlobTxs(encConf.TxConfig.TxEncoder(), 4, 1000)).ToSliceOfBytes(),
		}
	}

	type test struct {
		name           string
		input          *core.Data
		mutator        func(*core.Data)
		expectedResult abci.ResponseProcessProposal_Result
	}

	tests := []test{
		{
			name:           "valid untouched data",
			input:          validData(),
			mutator:        func(d *core.Data) {},
			expectedResult: abci.ResponseProcessProposal_ACCEPT,
		},
		{
			name:  "removed first blob",
			input: validData(),
			mutator: func(d *core.Data) {
				d.Blobs = d.Blobs[1:]
			},
			expectedResult: abci.ResponseProcessProposal_REJECT,
		},
		{
			name:  "added an extra blob",
			input: validData(),
			mutator: func(d *core.Data) {
				d.Blobs = append(
					d.Blobs,
					core.Blob{NamespaceId: []byte{1, 2, 3, 4, 5, 6, 7, 8}, Data: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
				)
			},
			expectedResult: abci.ResponseProcessProposal_REJECT,
		},
		{
			name:  "modified a blob",
			input: validData(),
			mutator: func(d *core.Data) {
				d.Blobs[0] = core.Blob{NamespaceId: []byte{1, 2, 3, 4, 5, 6, 7, 8}, Data: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}}
			},
			expectedResult: abci.ResponseProcessProposal_REJECT,
		},
		{
			name:  "invalid namespace TailPadding",
			input: validData(),
			mutator: func(d *core.Data) {
				d.Blobs[0] = core.Blob{NamespaceId: appconsts.TailPaddingNamespaceID, Data: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}}
			},
			expectedResult: abci.ResponseProcessProposal_REJECT,
		},
		{
			name:  "invalid namespace TxNamespace",
			input: validData(),
			mutator: func(d *core.Data) {
				d.Blobs[0] = core.Blob{NamespaceId: appconsts.TxNamespaceID, Data: []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}}
			},
			expectedResult: abci.ResponseProcessProposal_REJECT,
		},
		{
			name:  "unsorted blobs",
			input: validData(),
			mutator: func(d *core.Data) {
				blob1, blob2, blob3 := d.Blobs[0], d.Blobs[1], d.Blobs[2]
				d.Blobs[0] = blob3
				d.Blobs[1] = blob1
				d.Blobs[2] = blob2
			},
			expectedResult: abci.ResponseProcessProposal_REJECT,
		},
		{
			name:  "modified hash based on sequence start indicator tampering",
			input: validData(),
			mutator: func(d *core.Data) {
				coreData, _ := coretypes.DataFromProto(d)
				dataSquare, _ := shares.Split(coreData, true)
				fmt.Printf("dataSquare[1] before: %v", dataSquare[1])
				dataSquare[1] = tamper(dataSquare[1])
				fmt.Printf("dataSquare[1] after: %v", dataSquare[1])
				eds, _ := da.ExtendShares(d.SquareSize, shares.ToBytes(dataSquare))
				dah := da.NewDataAvailabilityHeader(eds)
				d.Hash = dah.Hash()
			},
			expectedResult: abci.ResponseProcessProposal_REJECT,
		},
	}

	for _, tt := range tests {
		resp := testApp.PrepareProposal(abci.RequestPrepareProposal{
			BlockData: tt.input,
		})
		tt.mutator(resp.BlockData)
		res := testApp.ProcessProposal(abci.RequestProcessProposal{
			BlockData: resp.BlockData,
			Header: core.Header{
				DataHash: resp.BlockData.Hash,
			},
		})
		assert.Equal(t, tt.expectedResult, res.Result, tt.name)
	}
}

func TestProcessProposalWithParityShareNamespace(t *testing.T) {
	testApp, _ := testutil.SetupTestAppWithGenesisValSet()
	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)

	txs := coretypes.Txs(blobfactory.RandBlobTxs(encConf.TxConfig.TxEncoder(), 4, 1000)).ToSliceOfBytes()
	req := abci.RequestPrepareProposal{
		BlockData: &tmproto.Data{
			Txs: txs,
		},
	}

	resp := testApp.PrepareProposal(req)

	resp.BlockData.Blobs[0].NamespaceId = appconsts.ParitySharesNamespaceID

	input := abci.RequestProcessProposal{
		BlockData: resp.BlockData,
	}
	res := testApp.ProcessProposal(input)
	require.Equal(t, abci.ResponseProcessProposal_REJECT, res.Result)
}

// func TestProcessProposalWithSequenceStartModification(t *testing.T) {
// 	testApp, _ := testutil.SetupTestAppWithGenesisValSet()
// 	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)

// 	txs := coretypes.Txs(blobfactory.GenerateManyRawSendTxs(encConf.TxConfig, 10)).ToSliceOfBytes()
// 	req := abci.RequestPrepareProposal{
// 		BlockData: &tmproto.Data{
// 			Txs: txs,
// 		},
// 	}

// 	resp := testApp.PrepareProposal(req)
// 	input := abci.RequestProcessProposal{
// 		BlockData: resp.BlockData,
// 	}
// 	res := testApp.ProcessProposal(input)

// 	require.Equal(t, abci.ResponseProcessProposal_REJECT, res.Result)
// }

// tamper flips the sequence start indicator of the share provided
func tamper(share shares.Share) shares.Share {
	// the info byte is immediately after the namespace
	infoByteIndex := appconsts.NamespaceSize
	// the sequence start indicator is the last bit of the info byte so flip the
	// last bit
	share[infoByteIndex] = share[infoByteIndex] ^ 0x01

	return share
}
