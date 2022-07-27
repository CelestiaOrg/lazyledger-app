package types

import (
	"fmt"
	"math/bits"

	"github.com/celestiaorg/celestia-app/pkg/util"
	"github.com/celestiaorg/rsmt2d"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/pkg/consts"
	"github.com/tendermint/tendermint/pkg/wrapper"
	coretypes "github.com/tendermint/tendermint/types"
)

const (
	URLMsgWirePayForData = "/payment.MsgWirePayForData"
	URLMsgPayForData     = "/payment.MsgPayForData"
	ShareSize            = consts.ShareSize
	SquareSize           = consts.MaxSquareSize
	NamespaceIDSize      = consts.NamespaceSize
)

var _ sdk.Msg = &MsgPayForData{}

// Route fullfills the sdk.Msg interface
func (msg *MsgPayForData) Route() string { return RouterKey }

// Type fullfills the sdk.Msg interface
func (msg *MsgPayForData) Type() string {
	return URLMsgPayForData
}

// ValidateBasic fullfills the sdk.Msg interface by performing stateless
// validity checks on the msg that also don't require having the actual message
func (msg *MsgPayForData) ValidateBasic() error {
	// ensure that the namespace id is of length == NamespaceIDSize
	if nsLen := len(msg.GetMessageNamespaceId()); nsLen != NamespaceIDSize {
		return fmt.Errorf(
			"invalid namespace length: got %d wanted %d",
			nsLen,
			NamespaceIDSize,
		)
	}

	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return err
	}

	return nil
}

// GetSignBytes fullfills the sdk.Msg interface by reterning a deterministic set
// of bytes to sign over
func (msg *MsgPayForData) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners fullfills the sdk.Msg interface by returning the signer's address
func (msg *MsgPayForData) GetSigners() []sdk.AccAddress {
	address, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{address}
}

// BuildPayForDataTxFromWireTx creates an authsigning.Tx using data from the original
// MsgWirePayForData sdk.Tx and the signature provided. This is used while processing
// the MsgWirePayForDatas into Signed  MsgPayForData
func BuildPayForDataTxFromWireTx(
	origTx authsigning.Tx,
	builder sdkclient.TxBuilder,
	signature []byte,
	msg *MsgPayForData,
) (authsigning.Tx, error) {
	err := builder.SetMsgs(msg)
	if err != nil {
		return nil, err
	}
	builder = InheritTxConfig(builder, origTx)

	origSigs, err := origTx.GetSignaturesV2()
	if err != nil {
		return nil, err
	}
	if len(origSigs) != 1 {
		return nil, fmt.Errorf("unexpected number of signers: %d", len(origSigs))
	}

	newSig := signing.SignatureV2{
		PubKey: origSigs[0].PubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: signature,
		},
		Sequence: origSigs[0].Sequence,
	}

	err = builder.SetSignatures(newSig)
	if err != nil {
		return nil, err
	}

	return builder.GetTx(), nil
}

// CreateCommitment generates the commit bytes for a given message, namespace, and
// squaresize using a namespace merkle tree and the rules described at
// https://github.com/celestiaorg/celestia-specs/blob/master/src/rationale/message_block_layout.md#message-layout-rationale
func CreateCommitment(k uint64, namespace, message []byte) ([]byte, error) {
	msg := coretypes.Messages{
		MessagesList: []coretypes.Message{
			{
				NamespaceID: namespace,
				Data:        message,
			},
		},
	}

	// split into shares that are length delimited and include the namespace in
	// each share
	shares := msg.SplitIntoShares().RawShares()
	// if the number of shares is larger than that in the square, throw an error
	// note, we use k*k-1 here because at least a single share will be reserved
	// for the transaction paying for the message, therefore the max number of
	// shares a message can be is number of shares in square -1.
	if uint64(len(shares)) > (k*k)-1 {
		return nil, fmt.Errorf("message size exceeds max shares for square size %d: max %d taken %d", k, (k*k)-1, len(shares))
	}

	// organize shares for merkle mountain range
	heights := powerOf2MountainRange(uint64(len(shares)), k)
	leafSets := make([][][]byte, len(heights))
	cursor := uint64(0)
	for i, height := range heights {
		leafSets[i] = shares[cursor : cursor+height]
		cursor = cursor + height
	}

	// create the commits by pushing each leaf set onto an nmt
	subTreeRoots := make([][]byte, len(leafSets))
	for i, set := range leafSets {
		// create the nmt todo(evan) use nmt wrapper
		tree := wrapper.NewErasuredNamespacedMerkleTree(k)
		for _, leaf := range set {
			nsLeaf := append(make([]byte, 0), append(namespace, leaf...)...)
			// note: we're not concerned about adding the correct namespace to
			// erasure data since we're only dealing with original square data,
			// so we can push to the wrapped nmt using Axis and Cell == 0
			tree.Push(nsLeaf, rsmt2d.SquareIndex{Axis: 0, Cell: 0})
		}
		// add the root
		subTreeRoots[i] = tree.Root()
	}
	return merkle.HashFromByteSlices(subTreeRoots), nil
}

// powerOf2MountainRange returns the heights of the subtrees for binary merkle
// mountian range
func powerOf2MountainRange(l, k uint64) []uint64 {
	var output []uint64

	for l != 0 {
		switch {
		case l >= k:
			output = append(output, k)
			l = l - k
		case l < k:
			p := util.NextLowestPowerOf2(l)
			output = append(output, p)
			l = l - p
		}
	}

	return output
}

// DelimLen calculates the length of the delimiter for a given message size
func DelimLen(x uint64) int {
	return 8 - bits.LeadingZeros64(x)%8
}
