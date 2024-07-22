package pfm

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/celestiaorg/celestia-app/v2/app"
	utils "github.com/celestiaorg/celestia-app/v2/test/tokenfilter"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type PacketMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
}

type ForwardMetadata struct {
	Receiver       string        `json:"receiver"`
	Port           string        `json:"port"`
	Channel        string        `json:"channel"`
	Timeout        time.Duration `json:"timeout"`
	Retries        *uint8        `json:"retries,omitempty"`
	Next           *string       `json:"next,omitempty"`
	RefundSequence *uint64       `json:"refund_sequence,omitempty"`
}

func SetupTest(t *testing.T) (*ibctesting.Coordinator, *ibctesting.TestChain,
	*ibctesting.TestChain, *ibctesting.TestChain) {
	chains := make(map[string]*ibctesting.TestChain)
	coordinator := &ibctesting.Coordinator{
		T:           t,
		CurrentTime: time.Now(),
		Chains:      chains,
	}
	celestiaChain := utils.NewTestChain(t, coordinator, ibctesting.GetChainID(1))
	chainA := NewTestChain(t, coordinator, ibctesting.GetChainID(2))
	chainB := NewTestChain(t, coordinator, ibctesting.GetChainID(3))
	coordinator.Chains[ibctesting.GetChainID(1)] = celestiaChain
	coordinator.Chains[ibctesting.GetChainID(2)] = chainA
	coordinator.Chains[ibctesting.GetChainID(3)] = chainB
	return coordinator, chainA, celestiaChain, chainB
}

func NewTransferPaths(chain1, chain2, chain3 *ibctesting.TestChain) (*ibctesting.Path, *ibctesting.Path) {
	path1 := ibctesting.NewPath(chain1, chain2)
	path1.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path1.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path1.EndpointA.ChannelConfig.Version = types.Version
	path1.EndpointB.ChannelConfig.Version = types.Version
	path2 := ibctesting.NewPath(chain2, chain3)
	path2.EndpointA.ChannelConfig.PortID = ibctesting.TransferPort
	path2.EndpointB.ChannelConfig.PortID = ibctesting.TransferPort
	path2.EndpointA.ChannelConfig.Version = types.Version
	path2.EndpointB.ChannelConfig.Version = types.Version

	return path1, path2
}

// TestPacketForwardMiddlewareTransfer send a PFM transfer originating from Celestia to ChainA, then back to Celestia and finally to ChainB.
// It verifies that Celestia forwards the packet successfully, the balance of the sender account on Celestia decreases by the amount sent,
// and the balance of the receiver account on ChainB increases by the amount sent.
func TestPacketForwardMiddlewareTransfer(t *testing.T) {
	coordinator, chainA, celestiaChain, chainB := SetupTest(t)
	path1, path2 := NewTransferPaths(chainA, celestiaChain, chainB)

	coordinator.Setup(path1)
	coordinator.Setup(path2)

	celestiaApp := celestiaChain.App.(*app.App)
	originalCelestiaBal := celestiaApp.BankKeeper.GetBalance(celestiaChain.GetContext(), celestiaChain.SenderAccount.GetAddress(), sdk.DefaultBondDenom)

	// take half of the original balance
	amount := originalCelestiaBal.Amount.QuoRaw(2)
	timeoutHeight := clienttypes.NewHeight(1, 300)
	coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, amount)

	// Forward the packet to ChainB
	secondHopMetaData :=
		&PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: chainB.SenderAccount.GetAddress().String(),
				Channel:  path2.EndpointA.ChannelID,
				Port:     path2.EndpointA.ChannelConfig.PortID,
			},
		}

	nextBz, err := json.Marshal(secondHopMetaData)
	require.NoError(t, err)
	next := string(nextBz)

	// Send it back to Celestia
	firstHopMetaData :=
		&PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: celestiaChain.SenderAccount.GetAddress().String(),
				Channel:  path1.EndpointA.ChannelID,
				Port:     path1.EndpointA.ChannelConfig.PortID,
				Next:     &next,
			},
		}

	memo, err := json.Marshal(firstHopMetaData)
	require.NoError(t, err)

	// Transfer path: Celestia -> ChainA -> Celestia -> ChainB
	msg := types.NewMsgTransfer(path1.EndpointB.ChannelConfig.PortID, path1.EndpointB.ChannelID, coinToSendToB, celestiaChain.SenderAccount.GetAddress().String(), chainA.SenderAccount.GetAddress().String(), timeoutHeight, 0, string(memo))

	res, err := celestiaChain.SendMsgs(msg)
	require.NoError(t, err) 

	packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
	require.NoError(t, err) 

	err = ForwardPacket([]*ibctesting.Path{path1, path1, path2}, packet)
	require.NoError(t, err)

	celestiaBalanceAfter := celestiaApp.BankKeeper.GetBalance(celestiaChain.GetContext(), celestiaChain.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	require.Equal(t, originalCelestiaBal.Amount.Sub(amount), celestiaBalanceAfter.Amount)

	ibcDenomTrace := types.ParseDenomTrace(types.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), sdk.DefaultBondDenom))
	destinationBalanceAfter := chainB.App.(*SimApp).BankKeeper.GetBalance(chainB.GetContext(), chainB.SenderAccount.GetAddress(), ibcDenomTrace.IBCDenom())

	require.Equal(t, amount, destinationBalanceAfter.Amount)
}

func isPacketToEndpoint(endpoint *ibctesting.Endpoint, packet channeltypes.Packet) bool {
	pc := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(endpoint.Chain.GetContext(), packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	return bytes.Equal(pc, channeltypes.CommitPacket(endpoint.Chain.App.AppCodec(), packet))
}

// Submits packet to an endpoint and returns either the acknowledgement or another packet
func relayPacket(endpoint *ibctesting.Endpoint, packet channeltypes.Packet) (channeltypes.Packet, []byte, error) {
	if err := endpoint.UpdateClient(); err != nil {
		return channeltypes.Packet{}, nil, err
	}

	res, err := endpoint.RecvPacketWithResult(packet)
	if err != nil {
		return channeltypes.Packet{}, nil, err
	}

	ack, err := ibctesting.ParseAckFromEvents(res.GetEvents())
	if err != nil {
		packet, err = ibctesting.ParsePacketFromEvents(res.GetEvents())
		if err != nil {
			return channeltypes.Packet{}, nil, err
		}
		return packet, nil, nil
	}

	return packet, ack, nil
}

// ForwardPacket forwards a packet through a series of paths and routes the acknowledgement back
func ForwardPacket(paths []*ibctesting.Path, packet channeltypes.Packet) error {
	if len(paths) < 2 {
		return errors.New("path must have at least two hops to forward packet")
	}

	var (
		ack            []byte
		rewindEndpoint = make([]*ibctesting.Endpoint, len(paths))
		packets        = make([]channeltypes.Packet, len(paths))
	)

	// Relay the packet through the paths and store the packets and acknowledgements
	packets[0] = packet
	for idx, path := range paths {
		if isPacketToEndpoint(path.EndpointA, packets[idx]) {
			packet, packetAck, err := relayPacket(path.EndpointB, packets[idx])
			if err != nil {
				return err
			}
			if len(packetAck) == 0 {
				packets[idx+1] = packet
			} else {
				ack = packetAck
			}
			rewindEndpoint[idx] = path.EndpointA
		} else if isPacketToEndpoint(path.EndpointB, packets[idx]) {
			packet, packetAck, err := relayPacket(path.EndpointA, packets[idx])
			if err != nil {
				return err
			}
			if len(packetAck) == 0 {
				packets[idx+1] = packet
			} else {
				ack = packetAck
			}
			rewindEndpoint[idx] = path.EndpointB
		} else {
			return errors.New("packet is for neither endpoint A or endpoint B")
		}
	}

	if len(ack) == 0 {
		return errors.New("no acknowledgement received from the last packet")
	}

	// Now we route the acknowledgements back
	for i := len(rewindEndpoint) - 1; i >= 0; i-- {
		if err := rewindEndpoint[i].UpdateClient(); err != nil {
			return err
		}

		res, err := AcknowledgePacket(rewindEndpoint[i], packets[i], ack)
		if err != nil {
			return err
		}
		if i > 0 {
			ack, err = ibctesting.ParseAckFromEvents(res.GetEvents())
			if err != nil {
				return err
			}
		}
		rewindEndpoint[i].Chain.Coordinator.CommitBlock()
	}
	return nil
}

// AcknowledgePacket acknowledges a packet and returns the result
func AcknowledgePacket(endpoint *ibctesting.Endpoint, packet channeltypes.Packet, ack []byte) (*sdk.Result, error) {
	packetKey := host.PacketAcknowledgementKey(packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	proof, proofHeight := endpoint.Counterparty.QueryProof(packetKey)
	ackMsg := channeltypes.NewMsgAcknowledgement(packet, ack, proof, proofHeight, endpoint.Chain.SenderAccount.GetAddress().String())

	return endpoint.Chain.SendMsgs(ackMsg)
}
