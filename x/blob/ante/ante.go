package ante

import (
	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/pkg/shares"
	"github.com/celestiaorg/celestia-app/x/blob/types"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MinGasPFBDecorator helps to prevent a PFB from being included in a block
// but running out of gas in DeliverTx (effectively getting DA for free)
// This decorator should be run after any decorator that consumes gas.
type MinGasPFBDecorator struct {
	k BlobKeeper
}

func NewMinGasPFBDecorator(k BlobKeeper) MinGasPFBDecorator {
	return MinGasPFBDecorator{k}
}

// AnteHandle implemnts the AnteHandler interface. It checks to see
// if the transaction contains a MsgPayForBlobs and if so, checks that
// the transaction has allocated enough gas.
//
// TODO: We need to run this antehandler in process proposal
// however this will be a breaking change.
func (d MinGasPFBDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	var gasPerByte uint32
	txGas := ctx.GasMeter().GasRemaining()
	for _, m := range tx.GetMsgs() {
		// NOTE: here we assume only one PFB per transaction
		if pfb, ok := m.(*types.MsgPayForBlobs); ok {
			if gasPerByte == 0 {
				// lazily fetch the gas per byte param
				gasPerByte = d.k.GasPerBlobByte(ctx)
			}
			gasToConsume := gasToConsume(pfb, gasPerByte)
			if gasToConsume > txGas {
				return ctx, errors.Wrapf(sdkerrors.ErrInsufficientFee, "not enough gas to pay for blobs (minimum: %d, got: %d)", gasToConsume, txGas)
			}
		}
	}

	return next(ctx, tx, simulate)
}

func gasToConsume(pfb *types.MsgPayForBlobs, gasPerByte uint32) uint64 {
	var totalSharesUsed uint64
	for _, size := range pfb.BlobSizes {
		totalSharesUsed += uint64(shares.SparseSharesNeeded(size))
	}

	return totalSharesUsed * appconsts.ShareSize * uint64(gasPerByte)
}

type BlobKeeper interface {
	GasPerBlobByte(ctx sdk.Context) uint32
}
