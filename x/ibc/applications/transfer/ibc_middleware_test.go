package transfer_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/functionx/fx-core/v2/x/ibc/applications/transfer"

	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"

	_ "github.com/functionx/fx-core/v2/app"
	"github.com/functionx/fx-core/v2/x/ibc/applications/transfer/types"
	ibctesting "github.com/functionx/fx-core/v2/x/ibc/testing"
)

func (suite *TransferTestSuite) TestOnChanOpenInit() {
	var (
		channel *channeltypes.Channel
		path    *ibctesting.Path
		chanCap *capabilitytypes.Capability
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{

		{
			"success", func() {}, true,
		},
		{
			"max channels reached", func() {
				path.EndpointA.ChannelID = channeltypes.FormatChannelIdentifier(math.MaxUint32 + 1)
			}, false,
		},
		{
			"invalid order - ORDERED", func() {
				channel.Ordering = channeltypes.ORDERED
			}, false,
		},
		{
			"invalid port ID", func() {
				path.EndpointA.ChannelConfig.PortID = ibctesting.MockPort
			}, false,
		},
		{
			"invalid version", func() {
				channel.Version = "version"
			}, false,
		},
		{
			"capability already claimed", func() {
				err := suite.chainA.GetSimApp().ScopedTransferKeeper.ClaimCapability(suite.chainA.GetContext(), chanCap, host.ChannelCapabilityPath(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID))
				suite.Require().NoError(err)
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			path = NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)
			path.EndpointA.ChannelID = ibctesting.FirstChannelID

			counterparty := channeltypes.NewCounterparty(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
			channel = &channeltypes.Channel{
				State:          channeltypes.INIT,
				Ordering:       channeltypes.UNORDERED,
				Counterparty:   counterparty,
				ConnectionHops: []string{path.EndpointA.ConnectionID},
				Version:        types.Version,
			}

			module, _, err := suite.chainA.App.GetIBCKeeper().PortKeeper.LookupModuleByPort(suite.chainA.GetContext(), ibctesting.TransferPort)
			suite.Require().NoError(err)

			chanCap, err = suite.chainA.App.GetScopedIBCKeeper().NewCapability(suite.chainA.GetContext(), host.ChannelCapabilityPath(ibctesting.TransferPort, path.EndpointA.ChannelID))
			suite.Require().NoError(err)

			cbs, ok := suite.chainA.App.GetIBCKeeper().Router.GetRoute(module)
			suite.Require().True(ok)

			tc.malleate() // explicitly change fields in channel and testChannel

			err = cbs.OnChanOpenInit(suite.chainA.GetContext(), channel.Ordering, channel.GetConnectionHops(),
				path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, chanCap, channel.Counterparty, channel.GetVersion(),
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func (suite *TransferTestSuite) TestOnChanOpenTry() {
	var (
		channel             *channeltypes.Channel
		chanCap             *capabilitytypes.Capability
		path                *ibctesting.Path
		counterpartyVersion string
	)

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{

		{
			"success", func() {}, true,
		},
		{
			"max channels reached", func() {
				path.EndpointA.ChannelID = channeltypes.FormatChannelIdentifier(math.MaxUint32 + 1)
			}, false,
		},
		{
			"capability already claimed in INIT should pass", func() {
				err := suite.chainA.GetSimApp().ScopedTransferKeeper.ClaimCapability(suite.chainA.GetContext(), chanCap, host.ChannelCapabilityPath(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID))
				suite.Require().NoError(err)
			}, true,
		},
		{
			"invalid order - ORDERED", func() {
				channel.Ordering = channeltypes.ORDERED
			}, false,
		},
		{
			"invalid port ID", func() {
				path.EndpointA.ChannelConfig.PortID = ibctesting.MockPort
			}, false,
		},
		{
			"invalid counterparty version", func() {
				counterpartyVersion = "version"
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			path = NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)
			path.EndpointA.ChannelID = ibctesting.FirstChannelID

			counterparty := channeltypes.NewCounterparty(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
			channel = &channeltypes.Channel{
				State:          channeltypes.TRYOPEN,
				Ordering:       channeltypes.UNORDERED,
				Counterparty:   counterparty,
				ConnectionHops: []string{path.EndpointA.ConnectionID},
				Version:        types.Version,
			}
			counterpartyVersion = types.Version

			module, _, err := suite.chainA.App.GetIBCKeeper().PortKeeper.LookupModuleByPort(suite.chainA.GetContext(), ibctesting.TransferPort)
			suite.Require().NoError(err)

			chanCap, err = suite.chainA.App.GetScopedIBCKeeper().NewCapability(suite.chainA.GetContext(), host.ChannelCapabilityPath(ibctesting.TransferPort, path.EndpointA.ChannelID))
			suite.Require().NoError(err)

			cbs, ok := suite.chainA.App.GetIBCKeeper().Router.GetRoute(module)
			suite.Require().True(ok)

			tc.malleate() // explicitly change fields in channel and testChannel

			version, err := cbs.OnChanOpenTry(suite.chainA.GetContext(), channel.Ordering, channel.GetConnectionHops(),
				path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, chanCap, channel.Counterparty, counterpartyVersion,
			)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().Equal(types.Version, version)
			} else {
				suite.Require().Error(err)
				suite.Require().Equal("", version)
			}

		})
	}
}

func (suite *TransferTestSuite) TestOnChanOpenAck() {
	var counterpartyVersion string

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{

		{
			"success", func() {}, true,
		},
		{
			"invalid counterparty version", func() {
				counterpartyVersion = "version"
			}, false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			path := NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.SetupConnections(path)
			path.EndpointA.ChannelID = ibctesting.FirstChannelID
			counterpartyVersion = types.Version

			module, _, err := suite.chainA.App.GetIBCKeeper().PortKeeper.LookupModuleByPort(suite.chainA.GetContext(), ibctesting.TransferPort)
			suite.Require().NoError(err)

			cbs, ok := suite.chainA.App.GetIBCKeeper().Router.GetRoute(module)
			suite.Require().True(ok)

			tc.malleate() // explicitly change fields in channel and testChannel

			err = cbs.OnChanOpenAck(suite.chainA.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointA.Counterparty.ChannelID, counterpartyVersion)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

		})
	}
}

func TestParseIncomingTransferField(t *testing.T) {
	testCases := []struct {
		name                string
		input               string
		expThisChainAddress string
		expFinalDestination string
		expPort             string
		expChannel          string
		expPass             bool
	}{
		{
			name:    "error - no-forward error thisChainAddress",
			input:   "cosmos1av497q6kjky9j5v3z95668d57q9ha80p5fypd4",
			expPass: false,
		},
		{
			name:    "error - no-forward empty thisChainAddress",
			input:   "",
			expPass: false,
		},
		{
			name:    "error - forward empty thisChainAddress",
			input:   "|transfer/channel-0:cosmos1av497q6kjky9j5v3z95668d57q9ha80p5fypd4",
			expPass: false,
		},
		{
			name:    "error - forward empty destinationAddress",
			input:   "fx1av497q6kjky9j5v3z95668d57q9ha80pwe45qy|transfer/channel-0:",
			expPass: false,
		},
		{
			name:                "ok - no-forward",
			input:               "fx1av497q6kjky9j5v3z95668d57q9ha80pwe45qy",
			expPass:             true,
			expThisChainAddress: "fx1av497q6kjky9j5v3z95668d57q9ha80pwe45qy",
		},
		{
			name:                "ok - forward empty portID",
			input:               "fx1av497q6kjky9j5v3z95668d57q9ha80pwe45qy|/channel-0:cosmos1av497q6kjky9j5v3z95668d57q9ha80p5fypd4",
			expPass:             true,
			expThisChainAddress: "fx1av497q6kjky9j5v3z95668d57q9ha80pwe45qy",
			expPort:             "",
			expChannel:          "channel-0",
			expFinalDestination: "cosmos1av497q6kjky9j5v3z95668d57q9ha80p5fypd4",
		},
		{
			name:                "ok - forward empty channelID",
			input:               "fx1av497q6kjky9j5v3z95668d57q9ha80pwe45qy|transfer/:cosmos1av497q6kjky9j5v3z95668d57q9ha80p5fypd4",
			expPass:             true,
			expThisChainAddress: "fx1av497q6kjky9j5v3z95668d57q9ha80pwe45qy",
			expPort:             "transfer",
			expChannel:          "",
			expFinalDestination: "cosmos1av497q6kjky9j5v3z95668d57q9ha80p5fypd4",
		},
		{
			name:                "ok - forward",
			input:               "fx1av497q6kjky9j5v3z95668d57q9ha80pwe45qy|transfer/channel-0:cosmos1av497q6kjky9j5v3z95668d57q9ha80p5fypd4",
			expPass:             true,
			expThisChainAddress: "fx1av497q6kjky9j5v3z95668d57q9ha80pwe45qy",
			expPort:             "transfer",
			expChannel:          "channel-0",
			expFinalDestination: "cosmos1av497q6kjky9j5v3z95668d57q9ha80p5fypd4",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			thisChainAddress, finalDestination, port, channel, err := transfer.ParseIncomingTransferField(tc.input)
			if tc.expPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				return
			}

			require.EqualValues(t, tc.expThisChainAddress, thisChainAddress.String())
			require.EqualValues(t, tc.expFinalDestination, finalDestination)
			require.EqualValues(t, tc.expPort, port)
			require.EqualValues(t, tc.expChannel, channel)
		})
	}
}

func TestGetDenomByIBCPacket(t *testing.T) {
	testCases := []struct {
		name          string
		sourcePort    string
		sourceChannel string
		destPort      string
		destChannel   string
		packetDenom   string
		expDenom      string
	}{
		{
			name:          "source token FX",
			sourcePort:    "transfer",
			sourceChannel: "channel-0",
			destPort:      "transfer",
			destChannel:   "channel-1",
			packetDenom:   "transfer/channel-0/FX",
			expDenom:      "FX",
		},
		{
			name:          "source token - eth0x61CAf09780f6F227B242EA64997a36c94a40Aa3a",
			sourcePort:    "transfer",
			sourceChannel: "channel-0",
			destPort:      "transfer",
			destChannel:   "channel-1",
			packetDenom:   "transfer/channel-0/eth0x61CAf09780f6F227B242EA64997a36c94a40Aa3a",
			expDenom:      "eth0x61CAf09780f6F227B242EA64997a36c94a40Aa3a",
		},
		{
			name:          "dest token - atom",
			sourcePort:    "transfer",
			sourceChannel: "channel-0",
			destPort:      "transfer",
			destChannel:   "channel-1",
			packetDenom:   "atom",
			expDenom:      types.ParseDenomTrace(fmt.Sprintf("%s/%s/%s", "transfer", "channel-1", "atom")).IBCDenom(),
		},
		{
			name:          "dest token - ibc denom a->b  b->c",
			sourcePort:    "transfer",
			sourceChannel: "channel-0",
			destPort:      "transfer",
			destChannel:   "channel-1",
			packetDenom:   "transfer/channel-2/atom",
			expDenom:      types.ParseDenomTrace(fmt.Sprintf("%s/%s/%s", "transfer", "channel-1", "transfer/channel-2/atom")).IBCDenom(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualValue := transfer.GetDenomByIBCPacket(tc.sourcePort, tc.sourceChannel, tc.destPort, tc.destChannel, tc.packetDenom)
			require.EqualValues(t, tc.expDenom, actualValue)
		})
	}
}
