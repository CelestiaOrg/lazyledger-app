package app_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/celestiaorg/celestia-app/x/blob/types"
	blobtypes "github.com/celestiaorg/celestia-app/x/blob/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	core "github.com/tendermint/tendermint/proto/tendermint/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	coretypes "github.com/tendermint/tendermint/types"

	"github.com/celestiaorg/celestia-app/app"
	"github.com/celestiaorg/celestia-app/app/encoding"
	appns "github.com/celestiaorg/celestia-app/pkg/namespace"
	"github.com/celestiaorg/celestia-app/testutil"
	"github.com/celestiaorg/celestia-app/testutil/blobfactory"
	"github.com/celestiaorg/celestia-app/testutil/testfactory"
)

func TestProcessProposal(t *testing.T) {
	encConf := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	accounts := testfactory.GenerateAccounts(6)
	testApp, kr := testutil.SetupTestAppWithGenesisValSet(accounts...)
	infos := queryAccountInfo(testApp, accounts, kr)
	signer := types.GenerateKeyringSigner(t, accounts[0])

	// create 3 single blob blobTxs that are signed with valid account numbers
	// and sequences
	blobTxs := blobfactory.ManyMultiBlobTx(
		t,
		encConf.TxConfig.TxEncoder(),
		kr,
		testutil.ChainID,
		accounts[:3],
		infos[:3],
		blobfactory.NestedBlobs(
			t,
			appns.RandomBlobNamespaces(3),
			[][]int{{100}, {1000}, {420}},
		),
	)

	// create 3 MsgSend transactions that are signed with valid account numbers
	// and sequences
	sendTxs := testutil.SendTxsWithAccounts(
		t,
		testApp,
		encConf.TxConfig.TxEncoder(),
		kr,
		1000,
		accounts[0],
		accounts[len(accounts)-3:],
		"",
	)

	// block with all blobs included
	validData := func() *tmproto.Data {
		return &tmproto.Data{
			Txs: blobTxs,
		}
	}

	// create block data with a PFB that is not indexed and has no blob
	unindexedData := validData()
	blobtx := testutil.RandBlobTxsWithAccounts(
		t,
		testApp,
		encConf.TxConfig.TxEncoder(),
		kr,
		1000,
		2,
		false,
		"",
		accounts[:1],
	)[0]
	btx, _ := coretypes.UnmarshalBlobTx(blobtx)
	unindexedData.Txs = append(unindexedData.Txs, btx.Tx)

	// create block data with a tx that is random data, and therefore cannot be
	// decoded into an sdk.Tx
	undecodableData := validData()
	undecodableData.Txs = append(unindexedData.Txs, tmrand.Bytes(300))

	mixedData := validData()
	mixedData.Txs = append(mixedData.Txs, coretypes.Txs(sendTxs).ToSliceOfBytes()...)

	// create an invalid block by adding an otherwise valid PFB, but an invalid
	// signature since there's no account
	badSigPFBData := validData()
	badSigBlobTx := testutil.RandBlobTxsWithManualSequence(
		t,
		encConf.TxConfig.TxEncoder(),
		kr,
		1000,
		1,
		false,
		"",
		accounts[:1],
		420, 42,
	)[0]
	badSigPFBData.Txs = append(badSigPFBData.Txs, badSigBlobTx)

	invalidNamespaceData := validData()
	blobs := blobfactory.ManyRandBlobs(t, 100)
	invalidNamespaceTx := blobfactory.MultiBlobTxInvalidNamespace(t, encConf.TxConfig.TxEncoder(), signer, 0, 0, blobs...)
	invalidNamespaceData.Txs = append(invalidNamespaceData.Txs, invalidNamespaceTx)

	type test struct {
		name           string
		input          *core.Data
		mutator        func(*core.Data)
		expectedResult abci.ResponseProcessProposal_Result
	}
	// ns1 := appns.MustNewV0(bytes.Repeat([]byte{1}, appns.NamespaceVersionZeroIDSize))
	// explicitly ignore the error from appns.New because we know the input is
	// invalid because it doesn't contain the namespace verzion zero prefix
	// data := bytes.Repeat([]byte{1}, 13)

	tests := []test{
		// 	{
		// 		name:           "valid untouched data",
		// 		input:          validData(),
		// 		mutator:        func(d *core.Data) {},
		// 		expectedResult: abci.ResponseProcessProposal_ACCEPT,
		// 	},
		// 	{
		// 		name:  "removed first blob",
		// 		input: validData(),
		// 		mutator: func(d *core.Data) {
		// 			d.Blobs = d.Blobs[1:]
		// 		},
		// 		expectedResult: abci.ResponseProcessProposal_REJECT,
		// 	},
		// 	{
		// 		name:  "added an extra blob",
		// 		input: validData(),
		// 		mutator: func(d *core.Data) {
		// 			d.Blobs = append(
		// 				d.Blobs,
		// 				core.Blob{
		// 					NamespaceId:      ns1.ID,
		// 					Data:             data,
		// 					NamespaceVersion: uint32(ns1.Version),
		// 					ShareVersion:     uint32(appconsts.ShareVersionZero),
		// 				},
		// 			)
		// 		},
		// 		expectedResult: abci.ResponseProcessProposal_REJECT,
		// 	},
		// 	{
		// 		name:  "modified a blob",
		// 		input: validData(),
		// 		mutator: func(d *core.Data) {
		// 			d.Blobs[0] = core.Blob{
		// 				NamespaceId:      ns1.ID,
		// 				Data:             data,
		// 				NamespaceVersion: uint32(ns1.Version),
		// 				ShareVersion:     uint32(appconsts.ShareVersionZero),
		// 			}
		// 		},
		// 		expectedResult: abci.ResponseProcessProposal_REJECT,
		// 	},
		// 	{
		// 		name:  "invalid namespace TailPadding",
		// 		input: validData(),
		// 		mutator: func(d *core.Data) {
		// 			d.Blobs[0] = core.Blob{
		// 				NamespaceId:      appns.TailPaddingNamespace.ID,
		// 				Data:             data,
		// 				NamespaceVersion: uint32(appns.TailPaddingNamespace.Version),
		// 				ShareVersion:     uint32(appconsts.ShareVersionZero),
		// 			}
		// 		},
		// 		expectedResult: abci.ResponseProcessProposal_REJECT,
		// 	},
		// 	{
		// 		name:  "invalid namespace TxNamespace",
		// 		input: validData(),
		// 		mutator: func(d *core.Data) {
		// 			d.Blobs[0] = core.Blob{
		// 				NamespaceId:      appns.TxNamespace.ID,
		// 				Data:             data,
		// 				NamespaceVersion: uint32(appns.TxNamespace.Version),
		// 				ShareVersion:     uint32(appconsts.ShareVersionZero),
		// 			}
		// 		},
		// 		expectedResult: abci.ResponseProcessProposal_REJECT,
		// 	},
		// 	{
		// 		name:  "invalid namespace ParityShares",
		// 		input: validData(),
		// 		mutator: func(d *core.Data) {
		// 			d.Blobs[0] = core.Blob{
		// 				NamespaceId:      appns.ParitySharesNamespace.ID,
		// 				Data:             data,
		// 				NamespaceVersion: uint32(appns.ParitySharesNamespace.Version),
		// 				ShareVersion:     uint32(appconsts.ShareVersionZero),
		// 			}
		// 		},
		// 		expectedResult: abci.ResponseProcessProposal_REJECT,
		// 	},
		// {
		// 	name:  "invalid blob namespace",
		// 	input: validData(),
		// 	mutator: func(d *core.Data) {
		// 		d.Blobs[0] = core.Blob{
		// 			NamespaceId:      invalidNamespace.ID,
		// 			Data:             data,
		// 			ShareVersion:     uint32(appconsts.ShareVersionZero),
		// 			NamespaceVersion: uint32(invalidNamespace.Version),
		// 		}
		// 	},
		// 	expectedResult: abci.ResponseProcessProposal_REJECT,
		// },
		{
			name:  "invalid namespace in index wrapper tx",
			input: validData(),
			mutator: func(d *core.Data) {
				rawTx := d.Txs[0]
				wrappedTx, isWrapped := coretypes.UnmarshalIndexWrapper(rawTx)
				assert.True(t, isWrapped)

				encCfg := encoding.MakeConfig(app.ModuleEncodingRegisters...)
				sdkTx, err := encCfg.TxConfig.TxDecoder()(wrappedTx.Tx)
				assert.NoError(t, err)

				msgs := sdkTx.GetMsgs()
				assert.Len(t, msgs, 1)
				msg := msgs[0]
				msgPFB, ok := msg.(*blobtypes.MsgPayForBlobs)
				assert.True(t, ok)
				msgPFB.Namespaces[0] = bytes.Repeat([]byte{1}, 33)

				// TODO replace the sdkTx message with msgPFB
				signedTx, err := blobfactory.NewSignedTx(t, encCfg.TxConfig.TxEncoder(), msgPFB)
				assert.NoError(t, err)

				newSdkTx, err := encCfg.TxConfig.TxEncoder()(signedTx)
				assert.NoError(t, err)

				newWrappedTx, err := coretypes.MarshalIndexWrapper(newSdkTx, wrappedTx.ShareIndexes...)
				assert.NoError(t, err)

				d.Txs[0] = newWrappedTx
				assert.NotEqual(t, rawTx, newWrappedTx)
			},
			expectedResult: abci.ResponseProcessProposal_REJECT,
		},
		// {
		// 	name:  "unsorted blobs",
		// 	input: validData(),
		// 	mutator: func(d *core.Data) {
		// 		blob1, blob2, blob3 := d.Blobs[0], d.Blobs[1], d.Blobs[2]
		// 		d.Blobs[0] = blob3
		// 		d.Blobs[1] = blob1
		// 		d.Blobs[2] = blob2
		// 	},
		// 	expectedResult: abci.ResponseProcessProposal_REJECT,
		// },
		// {
		// 	name:           "un-indexed PFB",
		// 	input:          unindexedData,
		// 	mutator:        func(d *core.Data) {},
		// 	expectedResult: abci.ResponseProcessProposal_REJECT,
		// },
		// {
		// 	name:           "undecodable tx",
		// 	input:          undecodableData,
		// 	mutator:        func(d *core.Data) {},
		// 	expectedResult: abci.ResponseProcessProposal_REJECT,
		// },
		// {
		// 	name:  "incorrectly sorted wrapped pfb's",
		// 	input: mixedData,
		// 	mutator: func(d *core.Data) {
		// 		// swap txs at index 3 and 4 (essentially swapping a PFB with a normal tx)
		// 		d.Txs[4], d.Txs[3] = d.Txs[3], d.Txs[4]
		// 	},
		// 	expectedResult: abci.ResponseProcessProposal_REJECT,
		// },
		// {
		// 	// while this test passes and the block gets rejected, it is getting
		// 	// rejected because the data root is different. We need to refactor
		// 	// prepare proposal to abstract functionality into a different
		// 	// function or be able to skip the filtering checks. TODO: perform
		// 	// the mentioned refactor and make it easier to create invalid
		// 	// blocks for testing.
		// 	name:  "included pfb with bad signature",
		// 	input: validData(),
		// 	mutator: func(d *core.Data) {
		// 		btx, _ := coretypes.UnmarshalBlobTx(badSigBlobTx)
		// 		d.Txs = append(d.Txs, btx.Tx)
		// 		d.Blobs = append(d.Blobs, deref(btx.Blobs)...)
		// 		sort.SliceStable(d.Blobs, func(i, j int) bool {
		// 			return bytes.Compare(d.Blobs[i].NamespaceId, d.Blobs[j].NamespaceId) < 0
		// 		})
		// 		// todo: replace the data root with an updated hash
		// 	},
		// 	expectedResult: abci.ResponseProcessProposal_REJECT,
		// },
		// {
		// 	name: "tampered sequence start",
		// 	input: &tmproto.Data{
		// 		Txs: coretypes.Txs(sendTxs).ToSliceOfBytes(),
		// 	},
		// 	mutator: func(d *tmproto.Data) {
		// 		bd, err := coretypes.DataFromProto(d)
		// 		require.NoError(t, err)

		// 		dataSquare, err := shares.Split(bd, true)
		// 		require.NoError(t, err)

		// 		b := shares.NewEmptyBuilder().ImportRawShare(dataSquare[1].ToBytes())
		// 		b.FlipSequenceStart()
		// 		updatedShare, err := b.Build()
		// 		require.NoError(t, err)
		// 		dataSquare[1] = *updatedShare

		// 		eds, err := da.ExtendShares(d.SquareSize, shares.ToBytes(dataSquare))
		// 		require.NoError(t, err)

		// 		dah := da.NewDataAvailabilityHeader(eds)
		// 		// replace the hash of the prepare proposal response with the hash of a data
		// 		// square with a tampered sequence start indicator
		// 		d.Hash = dah.Hash()
		// 	},
		// 	expectedResult: abci.ResponseProcessProposal_REJECT,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			assert.Equal(t, tt.expectedResult, res.Result, fmt.Sprintf("expected %v, got %v", tt.expectedResult, res.Result))
		})
	}
}

func deref[T any](s []*T) []T {
	t := make([]T, len(s))
	for i, ss := range s {
		t[i] = *ss
	}
	return t
}
