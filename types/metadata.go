package types

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func GetFXMetaData(denom string) banktypes.Metadata {
	return banktypes.Metadata{
		Description: "The native staking token of the Function X",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: 0,
			},
		},
		Base:    denom,
		Display: denom,
		Name:    "Function X",
		Symbol:  denom,
	}
}

func GetCrossChainMetadata(name, symbol string, decimals uint32, denom string) banktypes.Metadata {
	return banktypes.Metadata{
		Description: "The cross chain token of the Function X",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: 0,
			},
			{
				Denom:    symbol,
				Exponent: decimals,
			},
		},
		Base:    denom,
		Display: denom,
		Name:    name,
		Symbol:  symbol,
	}
}

var (
	/*
			//// example of origin denom
			{
				"description": "The native staking token of the Function X",
				"denom_units": [
					{
						"denom": "FX",
						"exponent": 0,
						"aliases": []
					}
				],
				"base": "FX",
				"display": "FX",
				"name": "Function X",
				"symbol": "FX"
			}
			//// example of other denom
			{
				"description":"The cross chain token of the Function X",
				"denom_units":[
					{
						"denom":"eth0x338E7A8687AdA7274Dc87C95D94f920d8F4185AE",
						"exponent":0,
						"aliases":[

						]
					},
					{
						"denom":"PUNDIX",
						"exponent":18,
						"aliases":[

						]
					}
				],
				"base":"eth0x338E7A8687AdA7274Dc87C95D94f920d8F4185AE",
				"display":"eth0x338E7A8687AdA7274Dc87C95D94f920d8F4185AE",
				"name":"Pundi X Token",
				"symbol":"PUNDIX"
			}
			//// example many to one denom
		    {
		        "description":"The cross chain token of the Function X",
		        "denom_units":[
		            {
		                "denom":"usdt",
		                "exponent":0,
		                "aliases":[
		                    "tronTR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
		                    "polygon0xc2132D05D31c914a87C6611C10748AEb04B58e8F"
		                ]
		            },
		            {
		                "denom":"USDT",
		                "exponent":6,
		                "aliases":[]
		            }
		        ],
		        "base":"usdt",
		        "display":"usdt",
		        "name":"Tether USD",
		        "symbol":"USDT"
		    }
	*/

	wfxMetadata = GetFXMetaData(DefaultDenom)

	testnetPUNDIXMetadata   = GetCrossChainMetadata("Pundi X Token", "PUNDIX", 18, "eth0xd9EEd31F5731DfC3Ca18f09B487e200F50a6343B")
	testnetEthUSDTMetadata  = GetCrossChainMetadata("Tether USD", "USDT", 6, "eth0xD69133f9A0206b3340d9622F2eBc4571022b3b5f")
	testnetUSDCMetadata     = GetCrossChainMetadata("USD Coin", "USDC", 18, "eth0xeC822cd1238d946Cf0f73be57359c5cAa5512a9D")
	testnetDAIMetadata      = GetCrossChainMetadata("DAI StableCoin", "DAI", 18, "eth0x2870405E4ABF9FcCDc93d9cC83c09788296d8354")
	testnetPURSEMetadata    = GetCrossChainMetadata("PURSE TOKEN", "PURSE", 18, "ibc/4757BC3AA2C696F7083C825BD3951AE3D1631F2A272EA7AFB9B3E1CCCA8560D4")
	testnetUSDJMetadata     = GetCrossChainMetadata("JUST Stablecoin v1.0", "USDJ", 18, "tronTLBaRhANQoJFTqre9Nf1mjuwNWjCJeYqUL")
	testnetUSDFMetadata     = GetCrossChainMetadata("FX USD", "USDF", 6, "tronTK1pM7NtkLohgRgKA6LeocW2znwJ8JtLrQ")
	testnetTronUSDTMetadata = GetCrossChainMetadata("Tether USD", "USDT", 6, "tronTXLAQ63Xg1NAzckPwKHvzw7CSEmLMEqcdj")
	testnetLINKMetadata     = GetCrossChainMetadata("ChainLink Token", "LINK", 18, "polygon0x326C977E6efc84E512bB9C30f76E30c160eD06FB")

	mainnetPUNDIXMetadata = GetCrossChainMetadata("Pundi X Token", "PUNDIX", 18, "eth0x0FD10b9899882a6f2fcb5c371E17e70FdEe00C38")
	mainnetPURSEMetadata  = GetCrossChainMetadata("PURSE TOKEN", "PURSE", 18, "ibc/F08B62C2C1BE9E52942617489CAB1E94537FE3849F8EEC910B142468C340EB0D")
)

func GetMetadata() []banktypes.Metadata {
	if TestnetChainId == ChainId() {
		return []banktypes.Metadata{wfxMetadata, testnetPUNDIXMetadata, testnetEthUSDTMetadata, testnetUSDCMetadata, testnetDAIMetadata,
			testnetPURSEMetadata, testnetUSDJMetadata, testnetUSDFMetadata, testnetTronUSDTMetadata, testnetLINKMetadata}
	} else {
		return []banktypes.Metadata{wfxMetadata, mainnetPUNDIXMetadata, mainnetPURSEMetadata}
	}
}
