package utils

import (
	"context"

	"github.com/celestiaorg/celestia-app/v2/test/util/genesis"
	"github.com/celestiaorg/celestia-app/v2/test/util/testnode"
	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	tmdb "github.com/tendermint/tm-db"
)

func StartNode(ctx context.Context, config *testnode.Config, multiplexer *Multiplexer, rootDir string) (cctx testnode.Context, err error) {
	basePath, err := genesis.InitFiles(config.TmConfig.RootDir, config.TmConfig, config.Genesis, 0)
	if err != nil {
		return testnode.Context{}, err
	}
	config.AppOptions.Set(flags.FlagHome, basePath)

	cometNode, app, err := newCometNode(&config.UniversalTestingConfig, multiplexer)
	if err != nil {
		return testnode.Context{}, err
	}

	cctx = testnode.NewContext(ctx, config.Genesis.Keyring(), config.TmConfig, config.Genesis.ChainID, config.AppConfig.API.Address)

	cctx, _, err = testnode.StartNode(cometNode, cctx)
	if err != nil {
		return testnode.Context{}, err
	}

	cctx, _, err = testnode.StartGRPCServer(app, config.AppConfig, cctx)
	if err != nil {
		return testnode.Context{}, err
	}

	_, err = testnode.StartAPIServer(app, *config.AppConfig, cctx)
	if err != nil {
		return testnode.Context{}, err
	}

	return cctx, nil
}

func newCometNode(config *testnode.UniversalTestingConfig, multiplexer *Multiplexer) (cometNode *node.Node, app servertypes.Application, err error) {
	logger := testnode.NewLogger(config)
	db, err := tmdb.NewGoLevelDB("application", config.TmConfig.DBDir())
	if err != nil {
		return nil, nil, err
	}
	app = config.AppCreator(logger, db, nil, config.AppOptions)
	nodeKey, err := p2p.LoadOrGenNodeKey(config.TmConfig.NodeKeyFile())
	if err != nil {
		return nil, nil, err
	}
	cometNode, err = node.NewNode(
		config.TmConfig,
		privval.LoadOrGenFilePV(config.TmConfig.PrivValidatorKeyFile(), config.TmConfig.PrivValidatorStateFile()),
		nodeKey,
		// newProxyClientCreator(multiplexer),
		proxy.NewLocalClientCreator(app),
		node.DefaultGenesisDocProviderFunc(config.TmConfig),
		node.DefaultDBProvider,
		node.DefaultMetricsProvider(config.TmConfig.Instrumentation),
		logger,
	)
	if err != nil {
		return nil, nil, err
	}
	return cometNode, app, err
}

func newProxyClientCreator(multiplexer *Multiplexer) proxy.ClientCreator {
	return proxy.NewLocalClientCreator(multiplexer)
}
