package testfactory

import (
	"bytes"
	"context"
	"sort"

	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	"github.com/celestiaorg/celestia-app/pkg/namespace"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"google.golang.org/grpc"
)

func Repeat[T any](s T, count int) []T {
	ss := make([]T, count)
	for i := 0; i < count; i++ {
		ss[i] = s
	}
	return ss
}

const BaseAccountDefaultBalance = int64(10000)

// GenerateRandNamespacedRawData returns random data of length count. Each chunk
// of random data is of size shareSize and is prefixed with a random blob
// namespace.
func GenerateRandNamespacedRawData(count int) (result [][]byte) {
	for i := 0; i < count; i++ {
		rawData := tmrand.Bytes(appconsts.ShareSize)
		namespace := namespace.RandomBlobNamespace().Bytes()
		copy(rawData, namespace)
		result = append(result, rawData)
	}

	sortByteArrays(result)
	return result
}

func sortByteArrays(src [][]byte) {
	sort.Slice(src, func(i, j int) bool { return bytes.Compare(src[i], src[j]) < 0 })
}

func RandomAccountNames(count int) []string {
	accounts := make([]string, 0, count)
	for i := 0; i < count; i++ {
		accounts = append(accounts, tmrand.Str(10))
	}
	return accounts
}

func GenerateAccounts(count int) []string {
	accs := make([]string, count)
	for i := 0; i < count; i++ {
		accs[i] = tmrand.Str(20)
	}
	return accs
}

// NewBaseAccount creates a new base account.
// If an empty string is passed as a name, a random one will be generated and used.
//
// It takes a keyring and a name as its parameters.
// It returns a BaseAccount and a slice of sdk Coins with the default bond denom.
func NewBaseAccount(kr keyring.Keyring, name string) (*authtypes.BaseAccount, sdk.Coins) {
	if name == "" {
		name = tmrand.Str(6)
	}
	rec, _, err := kr.NewMnemonic(name, keyring.English, "", "", hd.Secp256k1)
	if err != nil {
		panic(err)
	}
	addr, err := rec.GetAddress()
	if err != nil {
		panic(err)
	}
	origCoins := sdk.Coins{sdk.NewInt64Coin(appconsts.BondDenom, BaseAccountDefaultBalance)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)
	return bacc, origCoins
}

func GetValidators(grpcConn *grpc.ClientConn) (stakingtypes.Validators, error) {
	scli := stakingtypes.NewQueryClient(grpcConn)
	vres, err := scli.Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{})
	if err != nil {
		return stakingtypes.Validators{}, err
	}
	return vres.Validators, nil
}

func GetAccountDelegations(grpcConn *grpc.ClientConn, address string) (stakingtypes.DelegationResponses, error) {
	cli := stakingtypes.NewQueryClient(grpcConn)
	res, err := cli.DelegatorDelegations(context.Background(),
		&stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: address})
	if err != nil {
		return nil, err
	}

	return res.DelegationResponses, nil
}

func GetAccountSpendableBalance(grpcConn *grpc.ClientConn, address string) (balances sdk.Coins, err error) {
	cli := banktypes.NewQueryClient(grpcConn)
	res, err := cli.SpendableBalances(
		context.Background(),
		&banktypes.QuerySpendableBalancesRequest{
			Address: address,
		},
	)
	if err != nil {
		return nil, err
	}
	return res.GetBalances(), nil
}

func GetRawAccountInfo(grpcConn *grpc.ClientConn, address string) ([]byte, error) {
	cli := authtypes.NewQueryClient(grpcConn)
	res, err := cli.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: address,
	})
	if err != nil {
		return nil, err
	}
	return res.Account.Value, nil
}
