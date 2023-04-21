package version

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChainVersionConfig(t *testing.T) {
	input := map[uint64]int64{
		1: 0,
		2: 10,
		3: 20,
	}
	vg, err := NewChainVersionConfig(input)
	require.NoError(t, err)
	require.Equal(t, "v1", vg.GetVersion(0))
	require.Equal(t, "v1", vg.GetVersion(1))
	require.Equal(t, "v1", vg.GetVersion(9))
	require.Equal(t, "v2", vg.GetVersion(10))
	require.Equal(t, "v2", vg.GetVersion(11))
	require.Equal(t, "v2", vg.GetVersion(19))
	require.Equal(t, "v3", vg.GetVersion(20))
	require.Equal(t, "v3", vg.GetVersion(math.MaxInt64))
}

func Test_createRange(t *testing.T) {
	type test struct {
		name    string
		input   map[uint64]int64
		want    []HeightRange
		wantErr bool
	}

	tests := []test{
		{
			name: "valid",
			input: map[uint64]int64{
				1: 0,
				2: 10,
				3: 20,
			},
			want: []HeightRange{
				{
					Start:   0,
					End:     9,
					Version: 1,
				},
				{
					Start:   10,
					End:     19,
					Version: 2,
				},
				{
					Start:   20,
					End:     math.MaxInt64, // the end height should be the max uint64
					Version: 3,
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := createRange(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
			// double check that all ranges are contiguous
			for i := 0; i < len(got)-1; i++ {
				require.Equal(t, got[i].End, got[i+1].Start-1)
			}
		})
	}
}
