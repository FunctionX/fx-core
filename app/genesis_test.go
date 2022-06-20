package app

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	fxtypes "github.com/functionx/fx-core/types"
)

const genesisData = `{"auth":{"params":{"max_memo_characters":"256","tx_sig_limit":"7","tx_size_cost_per_byte":"10","sig_verify_cost_ed25519":"590","sig_verify_cost_secp256k1":"1000"},"accounts":[]},"authz":{"authorization":[]},"bank":{"params":{"send_enabled":[],"default_send_enabled":true},"balances":[{"address":"fx16n3lc7cywa68mg50qhp847034w88pntquxjmcz","coins":[{"denom":"FX","amount":"378600525462891000000000000"}]}],"supply":[{"denom":"FX","amount":"378604525462891000000000000"}],"denom_metadata":[{"description":"The native staking token of the Function X","denom_units":[{"denom":"FX","exponent":0,"aliases":[]}],"base":"FX","display":"FX","name":"Function X","symbol":"FX"}]},"bsc":{"params":{"gravity_id":"fx-bridge-bsc","average_block_time":"5000","external_batch_timeout":"86400000","average_external_block_time":"5000","signed_window":"20000","slash_fraction":"0.001000000000000000","oracle_set_update_power_change_percent":"0.100000000000000000","ibc_transfer_timeout_height":"20000","delegate_threshold":{"denom":"FX","amount":"10000000000000000000000"},"delegate_multiple":"10"},"last_observed_block_height":null,"OracleSet":[],"oracle":[],"unbatched_transfers":[],"batches":[],"bridge_token":[]},"capability":{"index":"1","owners":[]},"crisis":{"constant_fee":{"denom":"FX","amount":"13333000000000000000000"}},"crosschain":{},"distribution":{"params":{"community_tax":"0.400000000000000000","base_proposer_reward":"0.010000000000000000","bonus_proposer_reward":"0.040000000000000000","withdraw_addr_enabled":true},"fee_pool":{"community_pool":[]},"delegator_withdraw_infos":[],"previous_proposer":"","outstanding_rewards":[],"validator_accumulated_commissions":[],"validator_historical_rewards":[],"validator_current_rewards":[],"delegator_starting_infos":[],"validator_slash_events":[]},"erc20":{"params":{"enable_erc20":true,"enable_evm_hook":true,"ibc_timeout":"43200s"},"token_pairs":[]},"evidence":{"evidence":[]},"evm":{"accounts":[],"params":{"evm_denom":"FX","enable_create":true,"enable_call":true,"extra_eips":[],"chain_config":{"homestead_block":"0","dao_fork_block":"0","dao_fork_support":true,"eip150_block":"0","eip150_hash":"0x0000000000000000000000000000000000000000000000000000000000000000","eip155_block":"0","eip158_block":"0","byzantium_block":"0","constantinople_block":"0","petersburg_block":"0","istanbul_block":"0","muir_glacier_block":"0","berlin_block":"0","london_block":"0","arrow_glacier_block":"0","merge_fork_block":"0"},"reject_unprotected_tx":false}},"feegrant":{"allowances":[]},"feemarket":{"params":{"no_base_fee":false,"base_fee_change_denominator":8,"elasticity_multiplier":2,"enable_height":"0","base_fee":"500000000000","min_gas_price":"500000000000.000000000000000000","min_gas_multiplier":"0.000000000000000000"},"block_gas":"0"},"genutil":{"gen_txs":[]},"gov":{"starting_proposal_id":"1","deposits":[],"votes":[],"proposals":[],"deposit_params":{"min_deposit":[{"denom":"FX","amount":"10000000000000000000000"}],"max_deposit_period":"1209600s"},"voting_params":{"voting_period":"1209600s"},"tally_params":{"quorum":"0.400000000000000000","threshold":"0.500000000000000000","veto_threshold":"0.334000000000000000"}},"gravity":{"params":{"gravity_id":"fx-bridge-eth","contract_source_hash":"","bridge_eth_address":"","bridge_chain_id":"1","signed_valsets_window":"10000","signed_batches_window":"10000","signed_claims_window":"10000","target_batch_timeout":"43200000","average_block_time":"5000","average_eth_block_time":"15000","slash_fraction_valset":"0.001000000000000000","slash_fraction_batch":"0.001000000000000000","slash_fraction_claim":"0.001000000000000000","slash_fraction_conflicting_claim":"0.001000000000000000","unbond_slashing_valsets_window":"10000","ibc_transfer_timeout_height":"10000","valset_update_power_change_percent":"0.100000000000000000"},"last_observed_nonce":"0","valsets":[],"valset_confirms":[],"batches":[],"batch_confirms":[],"attestations":[],"delegate_keys":[],"erc20_to_denoms":[],"unbatched_transfers":[],"module_coins":[]},"ibc":{"client_genesis":{"clients":[],"clients_consensus":[],"clients_metadata":[],"params":{"allowed_clients":["07-tendermint"]},"create_localhost":false,"next_client_sequence":"0"},"connection_genesis":{"connections":[],"client_connection_paths":[],"next_connection_sequence":"0","params":{"max_expected_time_per_block":"30000000000"}},"channel_genesis":{"channels":[],"acknowledgements":[],"commitments":[],"receipts":[],"send_sequences":[],"recv_sequences":[],"ack_sequences":[],"next_channel_sequence":"0"}},"migrate":{},"mint":{"minter":{"inflation":"0.350000000000000000","annual_provisions":"0.000000000000000000"},"params":{"mint_denom":"FX","inflation_rate_change":"0.300000000000000000","inflation_max":"0.416762000000000000","inflation_min":"0.170000000000000000","goal_bonded":"0.510000000000000000","blocks_per_year":"6311520"}},"other":{},"params":{},"polygon":{"params":{"gravity_id":"fx-bridge-polygon","average_block_time":"5000","external_batch_timeout":"86400000","average_external_block_time":"5000","signed_window":"20000","slash_fraction":"0.001000000000000000","oracle_set_update_power_change_percent":"0.100000000000000000","ibc_transfer_timeout_height":"20000","delegate_threshold":{"denom":"FX","amount":"10000000000000000000000"},"delegate_multiple":"10"},"last_observed_block_height":null,"OracleSet":[],"oracle":[],"unbatched_transfers":[],"batches":[],"bridge_token":[]},"slashing":{"params":{"signed_blocks_window":"20000","min_signed_per_window":"0.050000000000000000","downtime_jail_duration":"600s","slash_fraction_double_sign":"0.050000000000000000","slash_fraction_downtime":"0.001000000000000000"},"signing_infos":[],"missed_blocks":[]},"staking":{"params":{"unbonding_time":"1814400s","max_validators":20,"max_entries":7,"historical_entries":20000,"bond_denom":"FX"},"last_total_power":"0","last_validator_powers":[],"validators":[],"delegations":[],"unbonding_delegations":[],"redelegations":[],"exported":false},"transfer":{"port_id":"transfer","denom_traces":[],"params":{"send_enabled":true,"receive_enabled":true}},"tron":{"params":{"gravity_id":"fx-bridge-tron","average_block_time":"5000","external_batch_timeout":"86400000","average_external_block_time":"5000","signed_window":"20000","slash_fraction":"0.001000000000000000","oracle_set_update_power_change_percent":"0.100000000000000000","ibc_transfer_timeout_height":"20000","delegate_threshold":{"denom":"FX","amount":"10000000000000000000000"},"delegate_multiple":"10"},"last_observed_block_height":null,"OracleSet":[],"oracle":[],"unbatched_transfers":[],"batches":[],"bridge_token":[]},"upgrade":{},"vesting":{}}`

func TestNewDefaultGenesisByDenom(t *testing.T) {
	encodingConfig := MakeEncodingConfig()
	genAppState := NewDefAppGenesisByDenom(fxtypes.DefaultDenom, encodingConfig.Marshaler)

	genAppStateStr, err := json.Marshal(genAppState)
	assert.NoError(t, err)

	t.Log(string(genAppStateStr))
	assert.Equal(t, genesisData, string(genAppStateStr))
}
