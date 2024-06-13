package types

import (
	"cosmossdk.io/errors"
)

var (
	ErrInvalidVersion        = errors.Register(ModuleName, 1, "signalled version must be either the current version or one greater")
	ErrUpgradePending        = errors.Register(ModuleName, 2, "upgrade is already pending")
	ErrInvalidUpgradeVersion = errors.Register(ModuleName, 3, "invalid upgrade version")
)
