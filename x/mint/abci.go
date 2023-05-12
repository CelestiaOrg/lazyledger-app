package mint

import (
	"time"

	"github.com/celestiaorg/celestia-app/x/mint/keeper"
	"github.com/celestiaorg/celestia-app/x/mint/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker updates the inflation rate, annual provisions, and then mints
// the block provision for the current block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	maybeSetGenesisTime(ctx, k)
	maybeUpdateMinter(ctx, k)
	mintBlockProvision(ctx, k)
}

// maybeSetGenesisTime sets the genesis time if the current block height is 1.
func maybeSetGenesisTime(ctx sdk.Context, k keeper.Keeper) {
	if ctx.BlockHeight() == 1 {
		genesisTime := ctx.BlockTime()
		minter := k.GetMinter(ctx)
		minter.GenesisTime = &genesisTime
		k.SetMinter(ctx, minter)
	}
}

// maybeUpdateMinter updates the inflation rate and annual provisions if the
// inflation rate has changed.
func maybeUpdateMinter(ctx sdk.Context, k keeper.Keeper) {
	minter := k.GetMinter(ctx)
	newInflationRate := minter.CalculateInflationRate(ctx)
	if newInflationRate == minter.InflationRate {
		// The minter's InflationRate AnnualProvisions already reflect the
		// values for this year. Exit early because we don't need to update
		// them.
		return
	}
	minter.InflationRate = newInflationRate
	k.SetMinter(ctx, minter)

	totalSupply := k.StakingTokenSupply(ctx)
	minter.AnnualProvisions = minter.CalculateAnnualProvisions(totalSupply)
	k.SetMinter(ctx, minter)
}

// mintBlockProvision mints the block provision for the current block.
func mintBlockProvision(ctx sdk.Context, k keeper.Keeper) {
	minter := k.GetMinter(ctx)
	mintedCoin := minter.CalculateBlockProvision()
	mintedCoins := sdk.NewCoins(mintedCoin)

	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	err = k.SendCoinsToFeeCollector(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyInflationRate, minter.InflationRate.String()),
			sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
}
