package orchestrator

import (
	"context"
	"fmt"
	"strings"

	paytypes "github.com/celestiaorg/celestia-app/x/payment/types"
	"github.com/celestiaorg/celestia-app/x/qgb/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/bytes"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client/http"
	coretypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
)

type AppClient interface {
	SubscribeValset(ctx context.Context) (<-chan types.Valset, error)
	SubscribeDataCommitment(ctx context.Context) (<-chan ExtendedDataCommitment, error)
	BroadcastTx(ctx context.Context, msg sdk.Msg) error
	QueryDataCommitments(ctx context.Context, commit string) ([]types.MsgDataCommitmentConfirm, error)
}

type ExtendedDataCommitment struct {
	Commitment bytes.HexBytes
	Start, End int64
	Nonce      uint64
}

type appClient struct {
	tendermintRPC *http.HTTP
	qgbRPC        *grpc.ClientConn
	logger        tmlog.Logger
	signer        *paytypes.KeyringSigner
}

func NewAppClient(logger tmlog.Logger, keyringAccount, chainID, coreRPC, appRPC string) (AppClient, error) {
	trpc, err := http.New(coreRPC, "/websocket")
	if err != nil {
		return nil, err
	}

	qgbGRPC, err := grpc.Dial(appRPC, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// open a keyring using the configured settings
	// TODO: optionally ask for input for a password
	ring, err := keyring.New("orchestrator", "test", "", strings.NewReader(""))
	if err != nil {
		return nil, err
	}

	signer := paytypes.NewKeyringSigner(
		ring,
		keyringAccount,
		chainID,
	)

	return &appClient{
		tendermintRPC: trpc,
		qgbRPC:        qgbGRPC,
		logger:        logger,
		signer:        signer,
	}, nil
}

func (ac *appClient) SubscribeValset(ctx context.Context) (<-chan types.Valset, error) {
	valsets := make(chan types.Valset, 10)
	results, err := ac.tendermintRPC.Subscribe(ctx, "valset-changes", "tm.event='Tx' AND message.module='qgb'")
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(valsets)
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-results:
				attributes := ev.Events[types.EventTypeValsetRequest]
				for _, attr := range attributes {
					if attr != types.AttributeKeyNonce {
						continue
					}

					queryClient := types.NewQueryClient(ac.qgbRPC)

					lastValsetResp, err := queryClient.LastValsetRequests(ctx, &types.QueryLastValsetRequestsRequest{})
					if err != nil {
						ac.logger.Error(err.Error())
						return
					}

					// todo: double check that the first validator set is found
					if len(lastValsetResp.Valsets) < 1 {
						ac.logger.Error("no validator sets found")
						return
					}

					valset := lastValsetResp.Valsets[0]

					valsets <- valset
				}
			}
		}

	}()

	return valsets, nil
}

func (ac *appClient) SubscribeDataCommitment(ctx context.Context) (<-chan ExtendedDataCommitment, error) {
	dataCommitments := make(chan ExtendedDataCommitment)

	queryClient := types.NewQueryClient(ac.qgbRPC)

	resp, err := queryClient.Params(ctx, &types.QueryParamsRequest{})
	if err != nil {
		return nil, nil
	}

	params := resp.Params
	window := params.DataCommitmentWindow

	results, err := ac.tendermintRPC.Subscribe(ctx, "height", coretypes.EventQueryNewBlockHeader.String())
	if err != nil {
		return nil, nil
	}

	go func() {
		defer close(dataCommitments)

		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-results:
				eventDataHeader := ev.Data.(coretypes.EventDataNewBlockHeader)
				height := eventDataHeader.Header.Height
				// todo: refactor to ensure that no ranges of blocks are missed if the
				// parameters are changed
				if height%int64(window) != 0 {
					continue
				}

				// TODO: calculate start height some other way that can handle changes
				// in the data window param
				startHeight := height - int64(window)
				endHeight := height

				// create and send the data commitment
				dcResp, err := ac.tendermintRPC.DataCommitment(
					ctx,
					fmt.Sprintf("block.height >= %d AND block.height <= %d",
						startHeight,
						endHeight,
					),
				)
				if err != nil {
					ac.logger.Error(err.Error())
					continue
				}

				// TODO: store the nonce in the state somehwere, so that we don't have
				// to assume what the nonce is
				nonce := uint64(height) / window

				dataCommitments <- ExtendedDataCommitment{
					Commitment: dcResp.DataCommitment,
					Start:      startHeight,
					End:        endHeight,
					Nonce:      nonce,
				}

			}
		}

	}()

	return dataCommitments, nil
}

func (ac *appClient) BroadcastTx(ctx context.Context, msg sdk.Msg) error {
	err := ac.signer.QueryAccountNumber(ctx, ac.qgbRPC)
	if err != nil {
		return err
	}

	// TODO: update this api via https://github.com/celestiaorg/celestia-app/pull/187/commits/37f96d9af30011736a3e6048bbb35bad6f5b795c
	tx, err := ac.signer.BuildSignedTx(ac.signer.NewTxBuilder(), msg)
	if err != nil {
		return err
	}

	rawTx, err := ac.signer.EncodeTx(tx)
	if err != nil {
		return err
	}

	resp, err := paytypes.BroadcastTx(ctx, ac.qgbRPC, 1, rawTx)
	if err != nil {
		return err
	}

	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("failure to broadcast tx: %s", resp.TxResponse.Data)
	}

	return nil
}

func (ac *appClient) QueryDataCommitments(ctx context.Context, commit string) ([]types.MsgDataCommitmentConfirm, error) {
	queryClient := types.NewQueryClient(ac.qgbRPC)

	confirmsResp, err := queryClient.DataCommitmentConfirmsByCommitment(ctx, &types.QueryDataCommitmentConfirmsByCommitmentRequest{
		Commitment: commit,
	})
	if err != nil {
		return nil, err
	}

	return confirmsResp.Confirms, nil
}
