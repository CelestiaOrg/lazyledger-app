package testutil

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/celestiaorg/celestia-app/x/mint/client/cli"
	minttypes "github.com/celestiaorg/celestia-app/x/mint/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"

	appnetwork "github.com/celestiaorg/celestia-app/test/util/network"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up x/mint integration test suite")

	genesisState := s.cfg.GenesisState
	var mintData minttypes.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData))

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down x/mint integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) jsonArgs() []string {
	return []string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
}

func (s *IntegrationTestSuite) textArgs() []string {
	return []string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", tmcli.OutputFlag)}
}

// getGenesisTime returns the genesis time from the genesis state.
func (s *IntegrationTestSuite) getGenesisTime() *time.Time {
	genesisState := s.cfg.GenesisState
	var mintData minttypes.GenesisState
	s.Require().NoError(s.cfg.Codec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData))
	return mintData.GetMinter().GenesisTime
}

// TestGetCmdQueryInflationRate tests that the CLI query command for inflation
// rate returns the correct value. This test assumes that the initial inflation
// rate is 0.08.
func (s *IntegrationTestSuite) TestGetCmdQueryInflationRate() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "json output",
			args: s.jsonArgs(),
			want: `0.080000000000000000`,
		},
		{
			name: "text output",
			args: s.textArgs(),
			want: `0.080000000000000000`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryInflationRate()
			clientCtx := val.ClientCtx

			got, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.want, strings.TrimSpace(got.String()))
		})
	}
}

// TestGetCmdQueryAnnualProvisions tests that the CLI query command for annual-provisions
// returns the correct value. This test assumes that the initial inflation
// rate is 0.08 and the initial total supply is 500_000_000 utia.
//
// TODO assert that total supply is 500_000_000 utia.
func (s *IntegrationTestSuite) TestGetCmdQueryAnnualProvisions() {
	val := s.network.Validators[0]

	testCases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "json output",
			args: s.jsonArgs(),
			want: `40000000.000000000000000000`,
		},
		{
			name: "text output",
			args: s.textArgs(),
			want: `40000000.000000000000000000`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryAnnualProvisions()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.want, strings.TrimSpace(out.String()))
		})
	}
}

// TestGetCmdQueryGenesisTime tests that the CLI command for genesis time
// returns the same time that is set in the genesis state. The CLI command to
// query genesis time looks like: `celestia-appd query mint genesis-time`
func (s *IntegrationTestSuite) TestGetCmdQueryGenesisTime() {
	val := s.network.Validators[0]
	want := s.getGenesisTime().String()

	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "json output",
			args: s.jsonArgs(),
		},
		{
			name: "text output",
			args: s.textArgs(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdQueryGenesisTime()
			clientCtx := val.ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)

			trimmed := strings.TrimSpace(out.String())
			s.Require().Equal(want, trimmed)
		})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestIntegrationTestSuite(t *testing.T) {
	cfg := appnetwork.DefaultConfig()
	suite.Run(t, NewIntegrationTestSuite(cfg))
}
