package types_test

import (
	"github.com/celestiaorg/celestia-app/x/qgb/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenesisStateValidate(t *testing.T) {
	specs := map[string]struct {
		src    *types.GenesisState
		expErr bool
	}{
		"default params": {src: types.DefaultGenesis(), expErr: false},
		"empty params": {src: &types.GenesisState{
			Params: &types.Params{
				DataCommitmentWindow: 0,
			},
		}, expErr: true},
		"invalid params: short block time": {src: &types.GenesisState{
			Params: &types.Params{
				DataCommitmentWindow: 10,
			},
		}, expErr: true},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.Validate()
			if spec.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
