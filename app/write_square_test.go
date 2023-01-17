package app

import (
	"bytes"
	"sort"
	"testing"

	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/celestiaorg/celestia-app/pkg/transaction"
	"github.com/celestiaorg/celestia-app/testutil/blobfactory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmlog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	coretypes "github.com/tendermint/tendermint/types"
)

func Test_finalizeLayout(t *testing.T) {
	ns1 := []byte{1, 1, 1, 1, 1, 1, 1, 1}
	ns2 := []byte{2, 2, 2, 2, 2, 2, 2, 2}
	ns3 := []byte{3, 3, 3, 3, 3, 3, 3, 3}

	type test struct {
		squareSize      uint64
		nonreserveStart int
		ptxs            []transaction.ParsedTx
		expectedIndexes [][]uint32
	}
	tests := []test{
		{
			squareSize:      4,
			nonreserveStart: 10,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1},
				[][]int{{1}},
			),
			expectedIndexes: [][]uint32{{10}},
		},
		{
			squareSize:      4,
			nonreserveStart: 10,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns1},
				blobfactory.Repeat([]int{100}, 2),
			),
			expectedIndexes: [][]uint32{{10}, {11}},
		},
		{
			squareSize:      4,
			nonreserveStart: 10,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns1, ns1, ns1, ns1, ns1, ns1, ns1, ns1, ns1},
				blobfactory.Repeat([]int{100}, 10),
			),
			expectedIndexes: [][]uint32{{10}, {11}, {12}, {13}, {14}, {15}},
		},
		{
			squareSize:      4,
			nonreserveStart: 7,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns1, ns1, ns1, ns1, ns1, ns1, ns1, ns1},
				blobfactory.Repeat([]int{100}, 9),
			),
			expectedIndexes: [][]uint32{{7}, {8}, {9}, {10}, {11}, {12}, {13}, {14}, {15}},
		},
		{
			squareSize:      4,
			nonreserveStart: 3,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns1, ns1},
				[][]int{{10000}, {10000}, {1000000}},
			),
			expectedIndexes: [][]uint32{},
		},
		{
			squareSize:      64,
			nonreserveStart: 32,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns1, ns1},
				[][]int{{1000}, {10000}, {100000}},
			),
			expectedIndexes: [][]uint32{
				// BlobMinSquareSize(2) = 2 so the first blob has to start at the
				// next multiple of 2 >= 32 which is 32. This blob occupies
				// shares 32 to 33.
				{32},
				// BlobMinSquareSize(20) = 8 so the second blob has to start at
				// the next multiple of 8 >= 34 which is 40. This blob occupies
				// shares 40 to 59.
				{40},
				// BlobMinSquareSize(199) = 16 so the third blob has to start at
				// the next multiple of 16 >= 60 which is 64. This blob occupies
				// shares 64 to 262.
				{64},
			},
		},
		{
			squareSize:      32,
			nonreserveStart: 32,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns2, ns1, ns1},
				[][]int{{100}, {100}, {100}},
			),
			expectedIndexes: [][]uint32{{34}, {32}, {33}},
		},
		{
			squareSize:      32,
			nonreserveStart: 32,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns2, ns1},
				[][]int{{100}, {1000}, {1000}},
			),
			expectedIndexes: [][]uint32{{32}, {36}, {34}},
		},
		{
			squareSize:      32,
			nonreserveStart: 32,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns2, ns1},
				[][]int{{100}, {1000}, {1000}},
			),
			expectedIndexes: [][]uint32{{32}, {36}, {34}},
		},
		{
			squareSize:      4,
			nonreserveStart: 2,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns3, ns2},
				[][]int{{100}, {1000}, {420}},
			),
			expectedIndexes: [][]uint32{{2}, {4}, {3}},
		},
		{
			squareSize:      4,
			nonreserveStart: 4,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns3, ns3, ns2},
				[][]int{{100}, {1000, 1000}, {420}},
			),
			expectedIndexes: [][]uint32{{4}, {6, 8}, {5}},
		},
		{
			squareSize:      4,
			nonreserveStart: 4,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns3, ns3, ns1, ns2, ns2},
				[][]int{{100}, {1400, 1000, 200, 200}, {420}},
			),
			expectedIndexes: [][]uint32{{4}, {8, 12, 5, 6}, {7}},
		},
		{
			squareSize:      4,
			nonreserveStart: 4,
			ptxs: generateParsedTxsWithNIDs(
				t,
				[][]byte{ns1, ns3, ns3, ns1, ns2, ns2},
				[][]int{{100}, {1000, 1400, 200, 200}, {420}},
			),
			expectedIndexes: [][]uint32{{4}, {8, 10, 5, 6}, {7}},
		},
	}
	for i, tt := range tests {
		res, blobs := finalizeLayout(tt.squareSize, tt.nonreserveStart, tt.ptxs)
		require.Equal(t, len(tt.expectedIndexes), len(res), i)
		for j, ptx := range res {
			assert.Equal(t, tt.expectedIndexes[j], ptx.ShareIndexes, i)
		}

		processedTxs := transaction.ProcessTxs(tmlog.NewNopLogger(), res)

		sort.SliceStable(blobs, func(i, j int) bool {
			return bytes.Compare(blobs[i].NamespaceId, blobs[j].NamespaceId) < 0
		})

		blockData := tmproto.Data{
			Txs:        processedTxs,
			Blobs:      blobs,
			SquareSize: tt.squareSize,
		}

		coreData, err := coretypes.DataFromProto(&blockData)
		require.NoError(t, err)

		_, err = shares.Split(coreData, true)
		require.NoError(t, err)
	}
}
