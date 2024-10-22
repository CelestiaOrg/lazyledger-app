package utils

import (
	"fmt"

	v1 "github.com/celestiaorg/celestia-app/v2/pkg/appconsts/v1"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
)

const (
	initialAppVersion = v1.Version
)

// Multiplexer implements the abci.Application interface
var _ abci.Application = (*Multiplexer)(nil)

// Multiplexer is used to switch between different versions of the application.
type Multiplexer struct {
	// logger is the logger used by the multiplexer
	logger log.Logger
	// db is the database used by the application
	db tmdb.DB
	// application is the current application
	application AppWithMigrations
	// currentAppVersion is the version of the application that is currently
	// running.
	currentAppVersion uint64
	// nextAppVersion is the version of the application that should be upgraded
	// to. This value only differs from currentAppVersion if the current height
	// is an upgrade height.
	nextAppVersion uint64
}

func NewMultiplexer(logger log.Logger, db tmdb.DB) *Multiplexer {
	application := NewAppV2(db)
	return &Multiplexer{
		logger:            logger,
		db:                db,
		application:       application,
		currentAppVersion: initialAppVersion,
		nextAppVersion:    initialAppVersion,
	}
}

//
// #region Consensus
//

func (m *Multiplexer) InitChain(request abci.RequestInitChain) abci.ResponseInitChain {
	m.logger.Debug(fmt.Sprintf("Multiplexer InitChain invoked with current app version %v request app version %v\n", m.currentAppVersion, request.ConsensusParams.Version.AppVersion))
	m.currentAppVersion = request.ConsensusParams.Version.AppVersion
	m.nextAppVersion = request.ConsensusParams.Version.AppVersion
	app := m.getCurrentApp()
	return app.InitChain(request)
}

func (m *Multiplexer) PrepareProposal(request abci.RequestPrepareProposal) abci.ResponsePrepareProposal {
	m.logger.Debug(fmt.Sprintf("Multiplexer PrepareProposal invoked with current app version %v\n", m.currentAppVersion))
	app := m.getCurrentApp()
	return app.PrepareProposal(request)
}

func (m *Multiplexer) ProcessProposal(request abci.RequestProcessProposal) abci.ResponseProcessProposal {
	m.logger.Debug(fmt.Sprintf("Multiplexer ProcessProposal invoked with current app version %v\n", m.currentAppVersion))
	app := m.getCurrentApp()
	return app.ProcessProposal(request)
}

func (m *Multiplexer) BeginBlock(request abci.RequestBeginBlock) abci.ResponseBeginBlock {
	m.logger.Debug(fmt.Sprintf("Multiplexer BeginBlock invoked with current app version %v\n", m.currentAppVersion))
	app := m.getCurrentApp()
	return app.BeginBlock(request)
}

func (m *Multiplexer) DeliverTx(request abci.RequestDeliverTx) abci.ResponseDeliverTx {
	m.logger.Debug(fmt.Sprintf("Multiplexer DeliverTx invoked with current app version %v\n", m.currentAppVersion))
	app := m.getCurrentApp()
	return app.DeliverTx(request)
}

func (m *Multiplexer) EndBlock(request abci.RequestEndBlock) abci.ResponseEndBlock {
	m.logger.Debug(fmt.Sprintf("Multiplexer EndBlock invoked with current app version %v height %v\n", m.currentAppVersion, request.Height))
	app := m.getCurrentApp()
	got := app.EndBlock(request)
	if got.ConsensusParamUpdates != nil && got.ConsensusParamUpdates.Version != nil {
		nextAppVersion := got.ConsensusParamUpdates.Version.AppVersion
		if m.nextAppVersion != nextAppVersion {
			fmt.Printf("Setting multiplexer next app version to %v\n", nextAppVersion)
			m.nextAppVersion = nextAppVersion
		}
	}
	return got
}

func (m *Multiplexer) Commit() abci.ResponseCommit {
	m.logger.Debug(fmt.Sprintf("Multiplexer Commit invoked with current app version %v\n", m.currentAppVersion))

	app := m.getCurrentApp()
	got := app.Commit()

	if m.isUpgradePending() {
		fmt.Printf("Multiplexer upgrade is pending from %v to %v\n", m.currentAppVersion, m.nextAppVersion)
		if m.nextAppVersion == 3 {
			m.application = NewAppV3(m.db)
		}
		m.currentAppVersion = m.nextAppVersion
		fmt.Printf("Multiplexer upgrade completed to %v\n", m.currentAppVersion)

		// appHash := m.RunMigrations()
		// got.Data = appHash
		return got
	}
	return got
}

//
// #region Mempool
//

func (m *Multiplexer) CheckTx(request abci.RequestCheckTx) abci.ResponseCheckTx {
	app := m.getCurrentApp()
	return app.CheckTx(request)
}

//
// #region Info
//

func (m *Multiplexer) Info(request abci.RequestInfo) abci.ResponseInfo {
	app := m.getCurrentApp()
	return app.Info(request)
}

func (m *Multiplexer) Query(request abci.RequestQuery) abci.ResponseQuery {
	app := m.getCurrentApp()
	return app.Query(request)
}

//
// #region Snapshot
//

func (m *Multiplexer) ApplySnapshotChunk(request abci.RequestApplySnapshotChunk) abci.ResponseApplySnapshotChunk {
	app := m.getCurrentApp()
	return app.ApplySnapshotChunk(request)
}

func (m *Multiplexer) ListSnapshots(request abci.RequestListSnapshots) abci.ResponseListSnapshots {
	app := m.getCurrentApp()
	return app.ListSnapshots(request)
}

func (m *Multiplexer) LoadSnapshotChunk(request abci.RequestLoadSnapshotChunk) abci.ResponseLoadSnapshotChunk {
	app := m.getCurrentApp()
	return app.LoadSnapshotChunk(request)
}

func (m *Multiplexer) OfferSnapshot(request abci.RequestOfferSnapshot) abci.ResponseOfferSnapshot {
	app := m.getCurrentApp()
	return app.OfferSnapshot(request)
}

//
// #region Other
//

func (m *Multiplexer) SetOption(request abci.RequestSetOption) abci.ResponseSetOption {
	app := m.getCurrentApp()
	return app.SetOption(request)
}

func (m *Multiplexer) RunMigrations() []byte {
	app := m.getCurrentApp()
	return app.RunMigrations()
}

//
// #region Private
//

func (m *Multiplexer) isUpgradePending() bool {
	return m.currentAppVersion != m.nextAppVersion
}

func (m *Multiplexer) getCurrentApp() AppWithMigrations {
	return m.application
}
