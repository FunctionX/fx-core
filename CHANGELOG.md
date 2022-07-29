<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Change log

## [Unreleased]

## [v2.2.1] - 2022-07-28

### Bug Fixes

* Fix transaction msg `MsgConvertCoin` `MsgConvertERC20` too much gas
* Fix crosschain to ethereum
* Fix tendermint subcommand

## [v2.2.0] - 2022-07-22

### Features

* Add query oracle reward in the crosschain module
* Check fxcored version when synchronizing blocks from scratch
* Add denom many to one support
* Update RegisterCoinProposal support denom many to one
* Add UpdateDenomAliasProposal and MsgConvertDenom

### Improvements

* Add ibc transfer route event
* Add gravity and crosschain attention event claimHash

## [v2.1.1] - 2022-07-11

### Bug Fixes

* Add support for the `x-crisis-skip-assert-invariants` CLI flag to the `start` command 
* CLI parse legacy proposal `InitCrossChainParamsProposal` failed
* Deleted Polygon(USDT) and Tron(USDT) contracts and metadata initialization during migration and upgrade

### Improvements

* Refactor gravity handle FxOriginatedTokenClaim

## [v2.1.0] - 2022-06-29

### Improvements

* Bump tendermint to v0.34.20.
* Bump cosmos-sdk to v0.45.5.
* The IBC version was upgraded from Cosmos-SDK/x/ibc to IBC-Go v3.1.0
* Added modules: feegrant、authz、feemarket、evm、erc20、migrate
* Migrate modules: auth、bank、distribution、gov、slashing、ibc、crosschain(bsc、polygon、tron)
* The previous Oracle deposit will be automatically delegated to the validator with the highest power value after the upgrade.  Oracle can modify the validator itself, requiring a manual delegate payment
* `MsgRequestBatch` add the field BaseFee
* Delete gravity and crosschain module ibc sequence key 
* Update crosschain params AverageBlockTime
* Bump ethermint to v0.16.1-fxcore-v2.0.0-rc3.
* Update block max gas to 30_000_000

### CLI Breaking Changes

* `fxcored unsafe-reset-all` command has been moved to the `fxcored tendermint` sub-command.
* `fxcored tendermint update-validator` command has been rename to the `fxcored tendermint unsafe-reset-priv-validator`
* `fxcored tendermint update-node-key` command has been rename to the `fxcored tendermint unsafe-reset-node-key`
* Remove bech32 PubKey support, Use pubkey in JSON format
* `fxcored debug addr` command has been moved and rename to the `fxcored keys prase`.
* `fxcored keys add` command flags `--algo` the default is eth_secp256k1; `--coin-type` the default is 60
* `fxcored keys add` command output add the EIP55 address
* Remove Cli flags `--gas-prices` default value
* Change Cli flags `--gas` default value with `80000`
* Change the `fxcored config` command output to lowercase      
* Remove `network` command

### API Breaking Changes

* Update FX metadata, delete `fx` denom
* Refactor `gravity` and `crosschain` module reset api routes
* The `gravity` and `crosschain` module add `ProjectedBatchTimeoutHeight` and `BridgeTokens` query api
* The `gravity`、`crosschain` and `other` reset API route add `/fx` prefix

### Features

* Support evm, enable ethereum compatibility
* Support EIP1559, the initial gas price is 500Gwei
* Account migrate, migrate fx prefix address to 0x prefix address, validator is not supported
* Add gRPC swagger-ui
* The `gravity/crosschain` module support targetIbc `0x` prefix
* Add `fxcored config update` command, only missing parts are added

### Bug Fixes

* Fix --node flag parsing. [issues#22](https://github.com/FunctionX/fx-core/issues/22)
* Fix --output flag parsing. [issues#34](https://github.com/FunctionX/fx-core/issues/34)
* Fix ibc router is not empty, receive address parse error

### Deprecated

* Remove `network` command
