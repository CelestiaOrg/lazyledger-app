package app

import (
	"bytes"
	"sort"

	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/pkg/da"
	"github.com/celestiaorg/celestia-app/pkg/inclusion"
	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/celestiaorg/celestia-app/x/blob/types"
	"github.com/celestiaorg/rsmt2d"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	coretypes "github.com/tendermint/tendermint/types"
)

const (
	rejectedPropBlockLog = "Rejected proposal block:"
)

func (app *App) ProcessProposal(req abci.RequestProcessProposal) abci.ResponseProcessProposal {
	// Check for blob inclusion:
	//  - each MsgPayForBlob included in a block should have a corresponding blob data in the block body
	//  - the commitment in each PFB should match the commitment for the shares that contain that blob data
	//  - there should be no unpaid-for data

	data, err := coretypes.DataFromProto(req.BlockData)
	if err != nil {
		logInvalidPropBlockError(app.Logger(), req.Header, "failure to unmarshal block data:", err)
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}

	if !sort.IsSorted(coretypes.BlobsByNamespace(data.Blobs)) {
		logInvalidPropBlock(app.Logger(), req.Header, "blobs are unsorted")
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}

	dataSquare, err := shares.Split(data, true)
	if err != nil {
		logInvalidPropBlockError(app.Logger(), req.Header, "failure to compute shares from block data:", err)
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}

	cacher := inclusion.NewSubtreeCacher(data.SquareSize)
	eds, err := rsmt2d.ComputeExtendedDataSquare(shares.ToBytes(dataSquare), appconsts.DefaultCodec(), cacher.Constructor)
	if err != nil {
		logInvalidPropBlockError(app.Logger(), req.Header, "failure to erasure the data square", err)
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}

	dah := da.NewDataAvailabilityHeader(eds)

	if !bytes.Equal(dah.Hash(), req.Header.DataHash) {
		logInvalidPropBlock(app.Logger(), req.Header, "proposed data root differs from calculated data root")
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}

	// iterate over all of the MsgPayForBlob transactions and ensure that their
	// commitments are subtree roots of the data root.
	commitmentCounter := 0
	for _, rawTx := range req.BlockData.Txs {
		malleatedTx, isMalleated := coretypes.UnmarshalIndexWrapper(rawTx)
		if !isMalleated {
			continue
		}

		tx, err := app.txConfig.TxDecoder()(malleatedTx.Tx)
		if err != nil {
			return abci.ResponseProcessProposal{
				Result: abci.ResponseProcessProposal_REJECT,
			}
		}

		var checkedTx bool

		for _, msg := range tx.GetMsgs() {
			if sdk.MsgTypeURL(msg) != types.URLMsgPayForBlob {
				continue
			}

			if !checkedTx {
				err = tx.ValidateBasic()
				if err != nil {
					app.Logger().Error("Tx including MsgPayForBlob is invalid")
					return abci.ResponseProcessProposal{
						Result: abci.ResponseProcessProposal_REJECT,
					}
				}
				checkedTx = true
			}

			pfb, ok := msg.(*types.MsgPayForBlob)
			if !ok {
				app.Logger().Error("Msg type does not match MsgPayForBlob URL")
				continue
			}

			if err = pfb.ValidateBasic(); err != nil {
				logInvalidPropBlockError(app.Logger(), req.Header, "invalid MsgPayForBlob", err)
				return abci.ResponseProcessProposal{
					Result: abci.ResponseProcessProposal_REJECT,
				}
			}

			signers := msg.GetSigners()
			if len(signers) != 1 {
				logInvalidPropBlockError(app.Logger(), req.Header, "cannot have multiple signers for MsgPayForBlob", err)
				return abci.ResponseProcessProposal{
					Result: abci.ResponseProcessProposal_REJECT,
				}
			}

			signer, err := sdk.AccAddressFromBech32(pfb.Signer)
			if err != nil {
				// this panic should be unreachable
				panic("signer address validation should not fail after basic validation has already been done")
			}

			if !bytes.Equal(signer, signers[0]) {
				logInvalidPropBlockError(app.Logger(), req.Header, "invalid signer for MsgPayForBlob", err)
				return abci.ResponseProcessProposal{
					Result: abci.ResponseProcessProposal_REJECT,
				}
			}

			commitment, err := inclusion.GetCommit(cacher, dah, int(malleatedTx.ShareIndex), shares.BlobSharesUsed(int(pfb.BlobSize)))
			if err != nil {
				logInvalidPropBlockError(app.Logger(), req.Header, "commitment not found", err)
				return abci.ResponseProcessProposal{
					Result: abci.ResponseProcessProposal_REJECT,
				}
			}

			if !bytes.Equal(pfb.ShareCommitment, commitment) {
				logInvalidPropBlock(app.Logger(), req.Header, "found commitment does not match user's")
				return abci.ResponseProcessProposal{
					Result: abci.ResponseProcessProposal_REJECT,
				}
			}

			commitmentCounter++
		}
	}

	// compare the number of MPFBs and blobs, if they aren't
	// identical, then we already know this block is invalid
	if commitmentCounter != len(req.BlockData.Blobs) {
		logInvalidPropBlock(app.Logger(), req.Header, "varying number of MsgPayForBlob and blobs in the same block")
		return abci.ResponseProcessProposal{
			Result: abci.ResponseProcessProposal_REJECT,
		}
	}
	return abci.ResponseProcessProposal{
		Result: abci.ResponseProcessProposal_ACCEPT,
	}
}

func logInvalidPropBlock(l log.Logger, h tmproto.Header, reason string) {
	l.Error(
		rejectedPropBlockLog,
		"reason",
		reason,
		"proposer",
		h.ProposerAddress,
	)
}

func logInvalidPropBlockError(l log.Logger, h tmproto.Header, reason string, err error) {
	l.Error(
		rejectedPropBlockLog,
		"reason",
		reason,
		"proposer",
		h.ProposerAddress,
		"err",
		err.Error(),
	)
}
