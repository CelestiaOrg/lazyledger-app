package app_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/celestiaorg/celestia-app/v2/app"
	"github.com/celestiaorg/celestia-app/v2/pkg/user"
	"github.com/celestiaorg/celestia-app/v2/test/util/testnode"
	"github.com/celestiaorg/celestia-app/v2/x/minfee"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func Test_testnode(t *testing.T) {
	t.Run("testnode can start a network with default chain ID", func(t *testing.T) {
		testnode.NewNetwork(t, testnode.DefaultConfig())
	})
	t.Run("testnode can query GRPC", func(t *testing.T) {
		config := testnode.DefaultConfig()
		cctx, _, _ := testnode.NewNetwork(t, config)

		queryClient := minfee.NewQueryClient(cctx.GRPCClient)
		got, err := queryClient.NetworkMinGasPrice(cctx.GoContext(), &minfee.QueryNetworkMinGasPrice{})
		require.NoError(t, err)
		fmt.Printf("got %v\n", got)

		cctx.WaitForHeight(1)
		// serviceClient := nodeservice.NewServiceClient(cctx.GRPCClient)
		// // TODO: need to figure out the correct contex tto pass here
		// foo, err := serviceClient.Config(cctx.GoContext(), &nodeservice.ConfigRequest{})
		// // got, err = queryServer.Config(cctx.GoContext(), &nodeservice.ConfigRequest{})
		// require.NoError(t, err)
		// fmt.Printf("foo %v\n", foo.MinimumGasPrice)

		bar, err := user.QueryMinimumGasPrice(cctx.GoContext(), cctx.GRPCClient)
		require.NoError(t, err)
		fmt.Printf("bar %v\n", bar)

	})
	t.Run("testnode can start with a custom MinGasPrice", func(t *testing.T) {
		wantMinGasPrice := float64(0.003)
		appConfig := testnode.DefaultAppConfig()
		appConfig.MinGasPrices = fmt.Sprintf("%v%s", wantMinGasPrice, app.BondDenom)
		config := testnode.DefaultConfig().WithAppConfig(appConfig)
		cctx, _, _ := testnode.NewNetwork(t, config)

		got, err := queryMinimumGasPrice(cctx.GoContext(), cctx.GRPCClient)
		require.NoError(t, err)
		assert.Equal(t, wantMinGasPrice, got)
	})
}

func queryMinimumGasPrice(ctx context.Context, grpcConn *grpc.ClientConn) (float64, error) {
	resp, err := nodeservice.NewServiceClient(grpcConn).Config(ctx, &nodeservice.ConfigRequest{})
	if err != nil {
		return 0, err
	}

	minGasCoins, err := sdktypes.ParseDecCoins(resp.MinimumGasPrice)
	if err != nil {
		return 0, err
	}

	minGasPrice := minGasCoins.AmountOf(app.BondDenom).MustFloat64()
	return minGasPrice, nil
}
