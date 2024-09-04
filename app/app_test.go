package app_test

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/celestiaorg/celestia-app/v2/app"
	"github.com/celestiaorg/celestia-app/v2/app/encoding"
	"github.com/celestiaorg/celestia-app/v2/x/minfee"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
)

func TestNew(t *testing.T) {
	logger := log.NewNopLogger()
	db := tmdb.NewMemDB()
	traceStore := &NoopWriter{}
	invCheckPeriod := uint(1)
	encodingConfig := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	upgradeHeight := int64(0)
	appOptions := NoopAppOptions{}

	got := app.New(logger, db, traceStore, invCheckPeriod, encodingConfig, upgradeHeight, appOptions)

	t.Run("initializes ICAHostKeeper", func(t *testing.T) {
		assert.NotNil(t, got.ICAHostKeeper)
	})
	t.Run("initializes ScopedICAHostKeeper", func(t *testing.T) {
		assert.NotNil(t, got.ScopedICAHostKeeper)
	})
	t.Run("initializes StakingKeeper", func(t *testing.T) {
		assert.NotNil(t, got.StakingKeeper)
	})
	t.Run("should have set StakingKeeper hooks", func(t *testing.T) {
		// StakingKeeper doesn't expose a GetHooks method so this checks if
		// hooks have been set by verifying the a subsequent call to SetHooks
		// will panic.
		assert.Panics(t, func() { got.StakingKeeper.SetHooks(nil) })
	})
	t.Run("should not have sealed the baseapp", func(t *testing.T) {
		assert.False(t, got.IsSealed())
	})
	t.Run("should have set the minfee key table", func(t *testing.T) {
		subspace := got.GetSubspace(minfee.ModuleName)
		hasKeyTable := subspace.HasKeyTable()
		assert.True(t, hasKeyTable)
	})
}

// Define a struct to match the structure of mocha-genesis.json if you want to unmarshal the JSON data
type GenesisFile struct {
	GenesisTime     string                 `json:"genesis_time"`
	ChainID         string                 `json:"chain_id"`
	InitialHeight   string                 `json:"initial_height"`
	ConsensusParams abci.ConsensusParams   `json:"consensus_params"` // Adjust type based on structure
	Validators      []abci.ValidatorUpdate `json:"validators"`
	AppState        json.RawMessage        `json:"app_state"` // Adjust type based on structure
}

func TestInitChainAgain(t *testing.T) {
}

func TestInitChain(t *testing.T) {
	logger := log.NewNopLogger()
	db := tmdb.NewMemDB()
	traceStore := &NoopWriter{}
	invCheckPeriod := uint(1)
	encodingConfig := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	upgradeHeight := int64(0)
	appOptions := NoopAppOptions{}

	type testCase struct {
		name         string
		request      abci.RequestInitChain
		wantResponse abci.ResponseInitChain
		wantPanic    bool
	}
	testCases := []testCase{
		{
			name:         "should panic if consensus params not set",
			request:      abci.RequestInitChain{},
			wantResponse: abci.ResponseInitChain{},
			wantPanic:    true,
		},
		// {
		// 	name:         "should not panic on Arabica genesis.json",
		// 	request:      getGenesis(t, "arabica-genesis.json"),
		// 	wantResponse: abci.ResponseInitChain{},
		// 	wantPanic:    false,
		// },
		{
			name:         "should not panic on Mocha genesis.json",
			request:      getGenesis(t, "mocha-genesis.json"),
			wantResponse: abci.ResponseInitChain{},
			wantPanic:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			application := app.New(logger, db, traceStore, invCheckPeriod, encodingConfig, upgradeHeight, appOptions)
			if tc.wantPanic {
				assert.Panics(t, func() { application.InitChain(tc.request) })
				return
			}
			got := application.InitChain(tc.request)
			assert.Equal(t, tc.wantResponse, got)
		})
	}

}

func getGenesis(t *testing.T, filename string) abci.RequestInitChain {
	file, err := os.Open(filename)
	require.NoError(t, err)
	defer file.Close()

	bytes, err := io.ReadAll(file)
	require.NoError(t, err)

	var genesis GenesisFile
	if err := json.Unmarshal(bytes, &genesis); err != nil {
		require.NoError(t, err)
	}

	return abci.RequestInitChain{
		Time:            parseTime(genesis.GenesisTime),
		ChainId:         genesis.ChainID,
		InitialHeight:   parseHeight(genesis.InitialHeight),
		Validators:      genesis.Validators,
		ConsensusParams: &abci.ConsensusParams{}, // TODO
		AppStateBytes:   genesis.AppState,
	}
}

func TestOfferSnapshot(t *testing.T) {
	logger := log.NewNopLogger()
	db := tmdb.NewMemDB()
	traceStore := &NoopWriter{}
	invCheckPeriod := uint(1)
	encodingConfig := encoding.MakeConfig(app.ModuleEncodingRegisters...)
	upgradeHeight := int64(0)
	appOptions := NoopAppOptions{}
	snapshotOption := getSnapshotOption(t)
	app := app.New(logger, db, traceStore, invCheckPeriod, encodingConfig, upgradeHeight, appOptions, snapshotOption)

	t.Run("should return ACCEPT", func(t *testing.T) {
		request := abci.RequestOfferSnapshot{
			Snapshot: &abci.Snapshot{
				Height:   0x1b07ec,
				Format:   0x2,
				Chunks:   0x1,
				Hash:     []uint8{0xaf, 0xa5, 0xe, 0x16, 0x45, 0x4, 0x2e, 0x45, 0xd3, 0x49, 0xdf, 0x83, 0x2a, 0x57, 0x9d, 0x64, 0xc8, 0xad, 0xa5, 0xb, 0x65, 0x1b, 0x46, 0xd6, 0xc3, 0x85, 0x6, 0x51, 0xd7, 0x45, 0x8e, 0xb8},
				Metadata: []uint8{0xa, 0x20, 0xaf, 0xa5, 0xe, 0x16, 0x45, 0x4, 0x2e, 0x45, 0xd3, 0x49, 0xdf, 0x83, 0x2a, 0x57, 0x9d, 0x64, 0xc8, 0xad, 0xa5, 0xb, 0x65, 0x1b, 0x46, 0xd6, 0xc3, 0x85, 0x6, 0x51, 0xd7, 0x45, 0x8e, 0xb8},
			},
			AppHash: []byte("apphash"),
		}
		want := abci.ResponseOfferSnapshot{Result: abci.ResponseOfferSnapshot_ACCEPT}
		got := app.OfferSnapshot(request)
		assert.Equal(t, want, got)
	})
}

func getSnapshotOption(t *testing.T) func(*baseapp.BaseApp) {
	snapshotDir := t.TempDir()
	snapshotDB, err := tmdb.NewDB("metadata", tmdb.GoLevelDBBackend, t.TempDir())
	require.NoError(t, err)
	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	require.NoError(t, err)
	interval := uint64(10)
	keepRecent := uint32(10)
	return baseapp.SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(interval, keepRecent))
}

// NoopWriter is a no-op implementation of a writer.
type NoopWriter struct{}

func (nw *NoopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// NoopAppOptions is a no-op implementation of servertypes.AppOptions.
type NoopAppOptions struct{}

func (nao NoopAppOptions) Get(string) interface{} {
	return nil
}

// Optional helper function to parse time from a string to time.Time (required by RequestInitChain)
func parseTime(timeStr string) time.Time {
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err) // In production code, handle this error appropriately
	}
	return parsedTime
}

// Optional helper function to parse height from a string to int64 (required by RequestInitChain)
func parseHeight(heightStr string) int64 {
	height, err := strconv.ParseInt(heightStr, 10, 64)
	if err != nil {
		panic(err) // In production code, handle this error appropriately
	}
	return height
}
