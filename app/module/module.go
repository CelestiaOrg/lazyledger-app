package module

import (
	"encoding/json"
	"fmt"
	"sort"

	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Manager defines a module manager that provides the high level utility for managing and executing
// operations for a group of modules
type Manager struct {
	versionedModules   map[uint64]map[string]sdkmodule.AppModule
	allModules         []sdkmodule.AppModule
	firstVersion       uint64
	lastVersion        uint64
	OrderInitGenesis   []string
	OrderExportGenesis []string
	OrderBeginBlockers []string
	OrderEndBlockers   []string
	OrderMigrations    []string
}

type VersionedModule struct {
	module                 sdkmodule.AppModule
	fromVersion, toVersion uint64
}

func NewVersionedModule(module sdkmodule.AppModule, fromVersion, toVersion uint64) VersionedModule {
	return VersionedModule{
		module:      module,
		fromVersion: fromVersion,
		toVersion:   toVersion,
	}
}

// NewManager creates a new Manager object
func NewManager(modules ...VersionedModule) (*Manager, error) {
	moduleMap := make(map[uint64]map[string]sdkmodule.AppModule)
	allModules := make([]sdkmodule.AppModule, len(modules))
	modulesStr := make([]string, 0, len(modules))
	firstVersion, lastVersion := uint64(0), uint64(0)
	for idx, module := range modules {
		if module.fromVersion == 0 {
			return nil, sdkerrors.ErrInvalidVersion.Wrapf("v0 is not a valid version for module %s", module.module.Name())
		}
		if module.fromVersion > module.toVersion {
			return nil, sdkerrors.ErrLogic.Wrapf("toVersion can not be less than fromVersion for module %s", module.module.Name())
		}
		for version := module.fromVersion; version <= module.toVersion; version++ {
			if moduleMap[version] == nil {
				moduleMap[version] = make(map[string]sdkmodule.AppModule)
			}
			moduleMap[version][module.module.Name()] = module.module
		}
		allModules[idx] = module.module
		modulesStr = append(modulesStr, module.module.Name())
		if firstVersion == 0 || module.fromVersion < firstVersion {
			firstVersion = module.fromVersion
		}
		if lastVersion == 0 || module.toVersion > lastVersion {
			lastVersion = module.toVersion
		}
	}

	return &Manager{
		versionedModules:   moduleMap,
		allModules:         allModules,
		firstVersion:       firstVersion,
		lastVersion:        lastVersion,
		OrderInitGenesis:   modulesStr,
		OrderExportGenesis: modulesStr,
		OrderBeginBlockers: modulesStr,
		OrderEndBlockers:   modulesStr,
	}, nil
}

// SetOrderInitGenesis sets the order of init genesis calls
func (m *Manager) SetOrderInitGenesis(moduleNames ...string) {
	m.assertNoForgottenModules("SetOrderInitGenesis", moduleNames)
	m.OrderInitGenesis = moduleNames
}

// SetOrderExportGenesis sets the order of export genesis calls
func (m *Manager) SetOrderExportGenesis(moduleNames ...string) {
	m.assertNoForgottenModules("SetOrderExportGenesis", moduleNames)
	m.OrderExportGenesis = moduleNames
}

// SetOrderBeginBlockers sets the order of set begin-blocker calls
func (m *Manager) SetOrderBeginBlockers(moduleNames ...string) {
	m.assertNoForgottenModules("SetOrderBeginBlockers", moduleNames)
	m.OrderBeginBlockers = moduleNames
}

// SetOrderEndBlockers sets the order of end-blocker calls
func (m *Manager) SetOrderEndBlockers(moduleNames ...string) {
	m.assertNoForgottenModules("SetOrderEndBlockers", moduleNames)
	m.OrderEndBlockers = moduleNames
}

// SetOrderMigrations sets the order of migrations to be run. If not set
// then migrations will be run with an order defined in `DefaultMigrationsOrder`.
func (m *Manager) SetOrderMigrations(moduleNames ...string) {
	m.assertNoForgottenModules("SetOrderMigrations", moduleNames)
	m.OrderMigrations = moduleNames
}

// RegisterInvariants registers all module invariants
func (m *Manager) RegisterInvariants(ir sdk.InvariantRegistry) {
	for _, module := range m.allModules {
		module.RegisterInvariants(ir)
	}
}

// RegisterServices registers all module services
func (m *Manager) RegisterServices(cfg sdkmodule.Configurator) {
	for _, module := range m.allModules {
		module.RegisterServices(cfg)
	}
}

// InitGenesis performs init genesis functionality for modules. Exactly one
// module must return a non-empty validator set update to correctly initialize
// the chain.
func (m *Manager) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, genesisData map[string]json.RawMessage, appVersion uint64) abci.ResponseInitChain {
	var validatorUpdates []abci.ValidatorUpdate
	ctx.Logger().Info("initializing blockchain state from genesis.json")
	modules, versionSupported := m.versionedModules[appVersion]
	if !versionSupported {
		panic(fmt.Sprintf("version %d not supported", appVersion))
	}
	for _, moduleName := range m.OrderInitGenesis {
		if genesisData[moduleName] == nil {
			continue
		}
		if modules[moduleName] == nil {
			continue
		}
		ctx.Logger().Debug("running initialization for module", "module", moduleName)

		moduleValUpdates := modules[moduleName].InitGenesis(ctx, cdc, genesisData[moduleName])

		// use these validator updates if provided, the module manager assumes
		// only one module will update the validator set
		if len(moduleValUpdates) > 0 {
			if len(validatorUpdates) > 0 {
				panic("validator InitGenesis updates already set by a previous module")
			}
			validatorUpdates = moduleValUpdates
		}
	}

	// a chain must initialize with a non-empty validator set
	if len(validatorUpdates) == 0 {
		panic(fmt.Sprintf("validator set is empty after InitGenesis, please ensure at least one validator is initialized with a delegation greater than or equal to the DefaultPowerReduction (%d)", sdk.DefaultPowerReduction))
	}

	return abci.ResponseInitChain{
		Validators: validatorUpdates,
	}
}

// ExportGenesis performs export genesis functionality for modules
func (m *Manager) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec, version uint64) map[string]json.RawMessage {
	genesisData := make(map[string]json.RawMessage)
	modules := m.versionedModules[version]
	for _, moduleName := range m.OrderExportGenesis {
		genesisData[moduleName] = modules[moduleName].ExportGenesis(ctx, cdc)
	}

	return genesisData
}

// assertNoForgottenModules checks that we didn't forget any modules in the
// SetOrder* functions.
func (m *Manager) assertNoForgottenModules(setOrderFnName string, moduleNames []string) {
	ms := make(map[string]bool)
	for _, m := range moduleNames {
		ms[m] = true
	}
	var missing []string
	for _, m := range m.allModules {
		if _, ok := ms[m.Name()]; !ok {
			missing = append(missing, m.Name())
		}
	}
	if len(missing) != 0 {
		panic(fmt.Sprintf(
			"%s: all modules must be defined when setting %s, missing: %v", setOrderFnName, setOrderFnName, missing))
	}
}

// MigrationHandler is the migration function that each module registers.
type MigrationHandler func(sdk.Context) error

// VersionMap is a map of moduleName -> version
type VersionMap map[string]uint64

// RunMigrations performs in-place store migrations for all modules. This
// function MUST be called when the state machine changes appVersion
func (m Manager) RunMigrations(ctx sdk.Context, cfg sdkmodule.Configurator, fromVersion, toVersion uint64) error {
	c, ok := cfg.(configurator)
	if !ok {
		return sdkerrors.ErrInvalidType.Wrapf("expected %T, got %T", configurator{}, cfg)
	}
	modules := m.OrderMigrations
	if modules == nil {
		modules = DefaultMigrationsOrder(m.ModuleNames(toVersion))
	}
	currentVersionModules, exists := m.versionedModules[fromVersion]
	if !exists {
		return sdkerrors.ErrInvalidVersion.Wrapf("version %d not supported", fromVersion)
	}
	nextVersionModules, exists := m.versionedModules[toVersion]
	if !exists {
		return sdkerrors.ErrInvalidVersion.Wrapf("version %d not supported", toVersion)
	}

	for _, moduleName := range modules {
		_, currentModuleExists := currentVersionModules[moduleName]
		nextModule, nextModuleExists := nextVersionModules[moduleName]

		// if the module exists for both upgrades
		if currentModuleExists && nextModuleExists {
			err := c.runModuleMigrations(ctx, moduleName, fromVersion, toVersion)
			if err != nil {
				return err
			}
		} else if !currentModuleExists && nextModuleExists {
			ctx.Logger().Info(fmt.Sprintf("adding a new module: %s", moduleName))
			moduleValUpdates := nextModule.InitGenesis(ctx, c.cdc, nextModule.DefaultGenesis(c.cdc))
			// The module manager assumes only one module will update the
			// validator set, and it can't be a new module.
			if len(moduleValUpdates) > 0 {
				return sdkerrors.ErrLogic.Wrap("validator InitGenesis update is already set by another module")
			}
		}
		// TODO: handle the case where a module is no longer supported (i.e. removed from the state machine)
	}

	return nil
}

// BeginBlock performs begin block functionality for all modules. It creates a
// child context with an event manager to aggregate events emitted from all
// modules.
func (m *Manager) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	ctx = ctx.WithEventManager(sdk.NewEventManager())

	modules := m.versionedModules[ctx.BlockHeader().Version.App]
	if modules == nil {
		panic(fmt.Sprintf("no modules for version %d", ctx.BlockHeader().Version.App))
	}
	for _, moduleName := range m.OrderBeginBlockers {
		module, ok := modules[moduleName].(sdkmodule.BeginBlockAppModule)
		if ok {
			module.BeginBlock(ctx, req)
		}
	}

	return abci.ResponseBeginBlock{
		Events: ctx.EventManager().ABCIEvents(),
	}
}

// EndBlock performs end block functionality for all modules. It creates a
// child context with an event manager to aggregate events emitted from all
// modules.
func (m *Manager) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	validatorUpdates := []abci.ValidatorUpdate{}

	modules := m.versionedModules[ctx.BlockHeader().Version.App]
	if modules == nil {
		panic(fmt.Sprintf("no modules for version %d", ctx.BlockHeader().Version.App))
	}
	for _, moduleName := range m.OrderEndBlockers {
		module, ok := modules[moduleName].(sdkmodule.EndBlockAppModule)
		if !ok {
			continue
		}
		moduleValUpdates := module.EndBlock(ctx, req)

		// use these validator updates if provided, the module manager assumes
		// only one module will update the validator set
		if len(moduleValUpdates) > 0 {
			if len(validatorUpdates) > 0 {
				panic("validator EndBlock updates already set by a previous module")
			}

			validatorUpdates = moduleValUpdates
		}
	}

	return abci.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
		Events:           ctx.EventManager().ABCIEvents(),
	}
}

// ModuleNames returns list of all module names, without any particular order.
func (m *Manager) ModuleNames(version uint64) []string {
	modules, ok := m.versionedModules[version]
	if !ok {
		return []string{}
	}

	ms := make([]string, len(modules))
	i := 0
	for m := range modules {
		ms[i] = m
		i++
	}
	return ms
}

func (m *Manager) SupportedVersions() []uint64 {
	output := make([]uint64, 0, m.lastVersion-m.firstVersion+1)
	for version := m.firstVersion; version <= m.lastVersion; version++ {
		if _, ok := m.versionedModules[version]; ok {
			output = append(output, version)
		}
	}
	return output
}

// DefaultMigrationsOrder returns a default migrations order: ascending alphabetical by module name,
// except x/auth which will run last, see:
// https://github.com/cosmos/cosmos-sdk/issues/10591
func DefaultMigrationsOrder(modules []string) []string {
	const authName = "auth"
	out := make([]string, 0, len(modules))
	hasAuth := false
	for _, m := range modules {
		if m == authName {
			hasAuth = true
		} else {
			out = append(out, m)
		}
	}
	sort.Strings(out)
	if hasAuth {
		out = append(out, authName)
	}
	return out
}
