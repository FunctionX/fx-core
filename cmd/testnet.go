package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"os"
	"path/filepath"

	fxCfg "github.com/functionx/fx-core/v2/server/config"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/functionx/fx-core/v2/app/cli"

	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	"github.com/functionx/fx-core/v2/app"

	fxtypes "github.com/functionx/fx-core/v2/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	flagNodeNamePrefix = "node-name-prefix"
	flagValidatorNum   = "validators"
	flagOutputDir      = "output-dir"
	flagNodeDaemonHome = "node-daemon-home"
	flagStartingIP     = "starting-ip"
	flagDockerImage    = "docker-image"
)

// testnetCmd get cmd to initialize all files for tendermint testnet and application
func testnetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for a fxchain testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:
	fxcored testnet -validators 4 -output-dir ./testnet --starting-ip 172.20.0.2
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			srvconfig.SetConfigTemplate(fxCfg.DefaultConfigTemplate())

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			serverCtx := server.GetServerContextFromCmd(cmd)

			outputDir := serverCtx.Viper.GetString(flagOutputDir)
			chainID := serverCtx.Viper.GetString(flags.FlagChainID)
			valNum := serverCtx.Viper.GetInt(flagValidatorNum)
			startingIPAddress := serverCtx.Viper.GetString(flagStartingIP)
			dockerImage := serverCtx.Viper.GetString(flagDockerImage)

			minGasPrices, _ := cmd.Flags().GetString(server.FlagMinGasPrices)
			err = initTestnet(
				clientCtx,
				serverCtx,
				outputDir,
				chainID,
				minGasPrices,
				serverCtx.Viper.GetString(flagNodeNamePrefix),
				serverCtx.Viper.GetString(flagNodeDaemonHome),
				startingIPAddress,
				serverCtx.Viper.GetString(flags.FlagKeyAlgorithm),
				serverCtx.Viper.GetString(cli.FlagDenom),
				valNum,
			)
			if err != nil {
				return err
			}
			if err = generateFxChainDockerComposeYml(valNum, chainID, startingIPAddress, dockerImage); err != nil {
				return err
			}
			return clientCtx.PrintString("Please run: docker-compose up -d\n")
		},
	}

	cmd.Flags().Int(flagValidatorNum, 4, "Number of validators to initialize the testnet with")
	cmd.Flags().String(flagOutputDir, "./testnet", "Directory to store initialization data for the testnet")
	cmd.Flags().String(flagNodeNamePrefix, "node", "Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, fxtypes.Name, "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIP, "172.20.0.2", "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().String(flags.FlagChainID, "fxcore", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(server.FlagMinGasPrices, fmt.Sprintf("4000000000000%s", fxtypes.DefaultDenom), "Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum")
	cmd.Flags().String(flags.FlagKeyAlgorithm, secp256k1.KeyType, "Key signing algorithm to generate keys for")
	cmd.Flags().String(cli.FlagDenom, fxtypes.DefaultDenom, "set the default coin denomination")
	cmd.Flags().String(flagDockerImage, "functionx/fx-core:latest", "set docker run image")

	return cmd
}

func initTestnet(
	clientCtx client.Context,
	serverCtx *server.Context,
	outputDir,
	chainID,
	minGasPrices,
	nodeNamePrefix,
	nodeDaemonHome,
	startingIP,
	algoStr,
	denom string,
	valNum int,
) error {
	fxAppConfig := fxCfg.DefaultConfig()
	fxAppConfig.MinGasPrices = minGasPrices
	fxAppConfig.API.Enable = true
	fxAppConfig.Telemetry.Enabled = true
	fxAppConfig.Telemetry.PrometheusRetentionTime = 60
	fxAppConfig.Telemetry.EnableHostnameLabel = false
	fxAppConfig.Telemetry.GlobalLabels = [][]string{{"chain_id", chainID}}
	fxAppConfig.BypassMinFee = fxCfg.BypassMinFee{
		MsgTypes: []string{
			sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}),
			sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}),
			sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}),
		},
	}

	var (
		genAccounts []authtypes.GenesisAccount
		genBalances []banktypes.Balance
		genFiles    = make([]string, 0)
		nodeIDs     = make([]string, valNum)
		valPubKeys  = make([]cryptotypes.PubKey, valNum)
	)

	// generate private keys, node IDs, and initial transactions
	for i := 0; i < valNum; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeNamePrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		serverCtx.Config.SetRoot(nodeDir)

		if err := os.MkdirAll(filepath.Join(nodeDir, "config"), os.ModePerm); err != nil {
			return err
		}

		ip, err := getIP(i, startingIP)
		if err != nil {
			return err
		}

		nodeIDs[i], valPubKeys[i], err = genutil.InitializeNodeValidatorFiles(serverCtx.Config)
		if err != nil {
			return err
		}

		memo := fmt.Sprintf("%s@%s:26656", nodeIDs[i], ip)
		genFiles = append(genFiles, serverCtx.Config.GenesisFile())

		kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, nodeDir, bufio.NewReader(os.Stdin))
		if err != nil {
			return err
		}

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(algoStr, keyringAlgos)
		if err != nil {
			return err
		}
		valAddr, mnemonic, err := testutil.GenerateSaveCoinKey(kb, nodeDirName, "", true, algo)
		if err != nil {
			return err
		}
		valKeyData, err := json.Marshal(map[string]string{"mnemonic": mnemonic, "algo": keyringAlgos.String()})
		if err != nil {
			return err
		}
		if err := writeFile(fmt.Sprintf("%v.json", "val-key"), nodeDir, valKeyData); err != nil {
			return err
		}

		amount := sdk.TokensFromConsensusPower(int64(40/valNum), sdk.DefaultPowerReduction)
		if i == 0 && 40%valNum != 0 {
			amount = sdk.TokensFromConsensusPower(int64(40/valNum+40%valNum), sdk.DefaultPowerReduction)
		}

		coins := sdk.Coins{sdk.NewCoin(denom, amount)}
		genBalances = append(genBalances, banktypes.Balance{Address: valAddr.String(), Coins: coins.Sort()})
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(valAddr, nil, 0, 0))

		createValMsg, err := stakingtypes.NewMsgCreateValidator(
			sdk.ValAddress(valAddr),
			valPubKeys[i],
			sdk.NewCoin(denom, sdk.TokensFromConsensusPower(1, sdk.DefaultPowerReduction)),
			stakingtypes.NewDescription(nodeDirName, "", "", "", ""),
			stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
			sdk.OneInt(),
		)
		if err != nil {
			return err
		}

		txBuilder := clientCtx.TxConfig.NewTxBuilder()
		if err := txBuilder.SetMsgs(createValMsg); err != nil {
			return err
		}

		txBuilder.SetMemo(memo)

		txFactory := tx.Factory{}
		txFactory = txFactory.
			WithChainID(chainID).
			WithMemo(memo).
			WithKeybase(kb).
			WithTxConfig(clientCtx.TxConfig)

		if err := tx.Sign(txFactory, nodeDirName, txBuilder, true); err != nil {
			return err
		}

		txBz, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return err
		}

		gentxsDir := filepath.Join(outputDir, "gentxs")
		if err := writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBz); err != nil {
			return err
		}

		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), fxAppConfig)
	}

	appGenState := app.NewDefAppGenesisByDenom(denom, clientCtx.Codec)
	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[authtypes.ModuleName], &authGenState)

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}

	authGenState.Accounts = accounts
	appGenState[authtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&authGenState)

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState)

	bankGenState.Balances = append(bankGenState.Balances, genBalances...)
	appGenState[banktypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&bankGenState)

	appGenStateJSON, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}

	// generate empty genesis files for each validator and save
	for _, gen := range genFiles {
		if err := genDoc.SaveAs(gen); err != nil {
			return err
		}
	}

	var appState json.RawMessage
	genTime := tmtime.Now()

	serverCtx.Config.P2P.AddrBookStrict = false
	serverCtx.Config.RPC.ListenAddress = "tcp://0.0.0.0:26657"

	for i := 0; i < valNum; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeNamePrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		genTxsDir := filepath.Join(outputDir, "gentxs")

		serverCtx.Config.Moniker = nodeDirName
		serverCtx.Config.SetRoot(nodeDir)

		ip, err := getIP(i, startingIP)
		if err != nil {
			return err
		}
		serverCtx.Config.P2P.ExternalAddress = fmt.Sprintf("%s:26656", ip)

		nodeID, valPubKey := nodeIDs[i], valPubKeys[i]
		initCfg := genutiltypes.NewInitConfig(chainID, genTxsDir, nodeID, valPubKey)

		genDoc, err := types.GenesisDocFromFile(serverCtx.Config.GenesisFile())
		if err != nil {
			return err
		}

		if appState == nil {
			appState, err = genutil.GenAppStateFromConfig(clientCtx.Codec, clientCtx.TxConfig, serverCtx.Config, initCfg, *genDoc, banktypes.GenesisBalancesIterator{})
			if err != nil {
				return err
			}
		}

		genFile := serverCtx.Config.GenesisFile()

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, chainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return clientCtx.PrintString(fmt.Sprintf("Successfully initialized %d node directories\n", valNum))
}

func getIP(i int, startingIPAddr string) (ip string, err error) {
	if len(startingIPAddr) == 0 {
		ip, err = server.ExternalIP()
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	return calculateIP(startingIPAddr, i)
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}

func calculateSubnet(ip string) string {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		panic(fmt.Errorf("%v: non ipv4 address", ip))
	}
	ipv4[3] = 0
	return ipv4.String()
}

func writeFile(name string, dir string, contents []byte) error {
	writePath := filepath.Join(dir)
	file := filepath.Join(writePath, name)

	err := tmos.EnsureDir(writePath, 0755)
	if err != nil {
		return err
	}

	err = tmos.WriteFile(file, contents, 0644)
	if err != nil {
		return err
	}

	return nil
}

const fxChainDockerComposeYmlTemplate = `version: '3'

services:
  {{range .Services}}
  {{.ContainerName}}:
    container_name: {{.ContainerName}}
    image: {{.Image}}
    command: start
    ports:{{range .Ports}}
      - "{{.}}"{{end}}
    volumes:
      - {{.Volumes}}
    networks:
      chain-net:
        ipv4_address: {{.IPv4Address}}
  {{end}}
networks:
  chain-net:
    name: {{.NetworkName}}
    driver: bridge
    ipam:
      driver: default
      config:
      - subnet: {{.Subnet}}
`

func generateFxChainDockerComposeYml(numValidators int, chainId, startingIPAddress, dockerImage string) error {
	data := map[string]interface{}{
		"NetworkName": fmt.Sprintf("%s-net", chainId),
		"Subnet":      fmt.Sprintf("%s/16", calculateSubnet(startingIPAddress)),
	}
	services := make([]map[string]interface{}, 0)
	for i := 0; i < numValidators; i++ {
		ip, err := getIP(i, startingIPAddress)
		if err != nil {
			return err
		}
		ports := []string{fmt.Sprintf("%d:26657", 26657+(i*2))}
		if i == 0 {
			ports = append(ports, "1317:1317")
			ports = append(ports, "9090:9090")
		}
		services = append(services, map[string]interface{}{
			"Image":         dockerImage,
			"ContainerName": fmt.Sprintf("%s-node%d", chainId, i),
			"IPv4Address":   ip,
			"Volumes":       fmt.Sprintf("./testnet/node%d/%s:/root/.%s", i, chainId, chainId),
			"Ports":         ports,
		})
	}
	data["Services"] = services
	tmpl, err := template.New("").Parse(fxChainDockerComposeYmlTemplate)
	if err != nil {
		return err
	}
	f, err := os.OpenFile("docker-compose.yml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, data)
}
