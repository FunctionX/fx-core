package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tmcfg "github.com/cometbft/cometbft/config"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/privval"
	tmtypes "github.com/cometbft/cometbft/types"
	tmversion "github.com/cometbft/cometbft/version"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdkserver "github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/functionx/fx-core/v7/client/grpc"
	"github.com/functionx/fx-core/v7/server"
	fxcfg "github.com/functionx/fx-core/v7/server/config"
	fxtypes "github.com/functionx/fx-core/v7/types"
)

const SPACE = "  "

type blockchain interface {
	GetChainId() (string, error)
	GetBlockHeight() (int64, error)
	GetSyncing() (bool, error)
	GetNodeInfo() (*tmservice.VersionInfo, error)
	CurrentPlan() (*upgradetypes.Plan, error)
	GetConsensusValidators() ([]*tmservice.Validator, error)
}

func doctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check your system for potential problems",
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := sdkserver.GetServerContextFromCmd(cmd)
			clientCtx := client.GetClientContextFromCmd(cmd)
			printPrompt()
			printOSInfo(serverCtx.Config.RootDir)
			printSelfInfo()
			chainId, err := checkGenesis(serverCtx.Config.GenesisFile())
			if err != nil {
				return err
			}
			if err = checkUpgradeInfo(serverCtx.Config.RootDir); err != nil {
				return err
			}
			bc, err := getBlockchain(clientCtx, serverCtx)
			if err != nil {
				return err
			}
			needUpgrade, err := checkBlockchainData(bc, chainId, serverCtx.Config.PrivValidatorKeyFile())
			if err != nil {
				return err
			}
			if err = checkAppConfig(serverCtx.Viper); err != nil {
				return err
			}
			if err = checkTmConfig(serverCtx.Config, needUpgrade); err != nil {
				return err
			}
			if err = checkCosmovisor(serverCtx.Config.RootDir, bc); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().String(flags.FlagHome, fxtypes.GetDefaultNodeHome(), "The application home directory")
	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
	cmd.Flags().String(flags.FlagGRPC, "", "the gRPC endpoint to use for this chain")
	cmd.Flags().Bool(flags.FlagGRPCInsecure, false, "allow gRPC over insecure channels, if not TLS the server must use TLS")
	return cmd
}

func printPrompt() {
	fmt.Printf(`
Please note that these warnings are just used to help the fxCore maintainers
If everything you use "fxcored" for is working fine: please don't worry; 
just ignore this. Thanks!
`)
	fmt.Printf("\n")
}

func printOSInfo(home string) {
	fmt.Printf("Computer Info:\n")
	fmt.Printf("%sOS/Arch: %s/%s\n", SPACE, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("%sCPU: %d\n", SPACE, runtime.NumCPU())
	memory, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	fmt.Printf("%sMemory Total: %.2f GB, Available: %.2f GB, UsedPercent: %.2f%%\n",
		SPACE, float64(memory.Total)/1024/1024/1024, float64(memory.Available)/1024/1024/1024, memory.UsedPercent)
	usage, err := disk.Usage(home)
	if err != nil {
		return
	}
	fmt.Printf("%sDisk Total: %.2f GB, Free: %.2f GB, UsedPercent: %.2f%%\n",
		SPACE, float64(usage.Total)/1024/1024/1024, float64(usage.Free)/1024/1024/1024, usage.UsedPercent)
	fmt.Printf("%s%sPath: %s\n", SPACE, SPACE, usage.Path)
}

func printSelfInfo() {
	fmt.Printf("fxcored Info:\n")
	info := version.NewInfo()
	fmt.Printf("%sVersion: %s\n", SPACE, info.Version)
	fmt.Printf("%sGit Commit: %s\n", SPACE, info.GitCommit)
	fmt.Printf("%sBuild Tags: %s\n", SPACE, info.BuildTags)
	fmt.Printf("%sGo Version: %s\n", SPACE, runtime.Version())
	fmt.Printf("%sCosmos SDK Version: %s\n", SPACE, info.CosmosSdkVersion)
	fmt.Printf("%sTendermint Version: %s\n", SPACE, tmversion.TMCoreSemVer)
}

func checkGenesis(genesisFile string) (string, error) {
	fmt.Printf("Genesis:\n")
	fmt.Printf("%sFile: %s\n", SPACE, genesisFile)

	doc, sha256, err := getGenesisDocAndSha256(genesisFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("%sWarning: Not found Genesis file!\n", SPACE)
			return "", nil
		}
		return "", err
	}
	if sha256 == fxtypes.MainnetGenesisHash && doc.ChainID == fxtypes.MainnetChainId {
		fmt.Printf("%sNetwork: Mainnet\n", SPACE)
		return doc.ChainID, nil
	} else if sha256 == fxtypes.TestnetGenesisHash && doc.ChainID == fxtypes.TestnetChainId {
		fmt.Printf("%sNetwork: Testnet\n", SPACE)
		return doc.ChainID, nil
	} else {
		fmt.Printf("%sWarning: Unknown Network!\n", SPACE)
		return "Unknown", nil
	}
}

func checkUpgradeInfo(homeDir string) error {
	file := filepath.Join(homeDir, "data", upgradetypes.UpgradeInfoFilename)
	data, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read upgrade info error: %s", err.Error())
	}
	upgradeInfo := new(upgradetypes.Plan)
	if err := json.Unmarshal(data, upgradeInfo); err != nil {
		return err
	}
	fmt.Printf("Upgrade Info:\n")
	fmt.Printf("%sFile: %s\n", SPACE, file)
	fmt.Printf("%sName: %s\n", SPACE, upgradeInfo.Name)
	fmt.Printf("%sHeight: %d\n", SPACE, upgradeInfo.Height)
	return nil
}

func getBlockchain(cliCtx client.Context, serverCtx *sdkserver.Context) (blockchain, error) {
	fmt.Printf("Blockchain Data:\n")
	grpcAddr := serverCtx.Viper.GetString(flags.FlagGRPC)
	newClient := grpc.NewClient(cliCtx)
	_, err := newClient.GetBlockHeight()
	if err == nil {
		fmt.Printf("%sRemote Node: %s %s\n", SPACE, cliCtx.NodeURI, grpcAddr)
		return newClient, nil
	}
	if len(grpcAddr) > 0 {
		return nil, err
	}
	if len(serverCtx.Config.RootDir) == 0 {
		fmt.Printf("%sWarning: Not found root dir\n", SPACE)
		return nil, nil
	}

	database, err := server.NewDatabase(serverCtx.Config, cliCtx.Codec)
	if err != nil {
		return nil, err
	}
	if database == nil {
		fmt.Printf("%sWarning: Not found data file!\n", SPACE)
		return nil, nil
	}
	fmt.Printf("%sData Dir: %s/data\n", SPACE, serverCtx.Config.RootDir)
	return database, nil
}

//gocyclo:ignore
func checkBlockchainData(bc blockchain, genesisId, privValidatorKeyFile string) (bool, error) {
	if bc == nil {
		return false, nil
	}
	chainId, err := bc.GetChainId()
	if err != nil {
		fmt.Printf("%sWarning: %s\n", SPACE, err.Error())
		return false, nil
	}
	if len(chainId) > 0 {
		fmt.Printf("%sChain ID: %s\n", SPACE, chainId)
	}
	blockHeight, err := bc.GetBlockHeight()
	if err != nil {
		return false, err
	}
	if blockHeight > 0 {
		fmt.Printf("%sBlock Height: %d\n", SPACE, blockHeight)
	}
	syncing, err := bc.GetSyncing()
	if err != nil {
		return false, nil
	}
	fmt.Printf("%sSyncing: %v\n", SPACE, syncing)
	pvKey := privval.FilePVKey{}
	keyJSONBytes, err := os.ReadFile(privValidatorKeyFile)
	if err != nil {
		return false, err
	}
	if err = tmjson.Unmarshal(keyJSONBytes, &pvKey); err != nil {
		return false, err
	}
	validators, err := bc.GetConsensusValidators()
	if err != nil {
		return false, err
	}
	for _, validator := range validators {
		if strings.EqualFold(
			sdk.ConsAddress(pvKey.Address.Bytes()).String(),
			validator.GetAddress(),
		) {
			fmt.Printf("%sNode Type: This node is a validator\n", SPACE)
		}
	}
	app, err := bc.GetNodeInfo()
	if err != nil {
		return false, err
	}
	fmt.Printf("%sNode Info: \n", SPACE)
	if len(app.Version) > 0 {
		fmt.Printf("%s%sVersion: %s\n", SPACE, SPACE, app.Version)
	}
	if len(app.GitCommit) > 0 {
		fmt.Printf("%s%sGit Commit: %s\n", SPACE, SPACE, app.GitCommit)
	}
	if len(app.BuildTags) > 0 {
		fmt.Printf("%s%sBuild Tags: %s\n", SPACE, SPACE, app.BuildTags)
	}
	if len(app.GoVersion) > 0 {
		fmt.Printf("%s%sGo Version: %s\n", SPACE, SPACE, app.GoVersion)
	}
	if len(app.CosmosSdkVersion) > 0 {
		fmt.Printf("%s%sCosmos SDK Version: %s\n", SPACE, SPACE, app.CosmosSdkVersion)
	}
	plan, err := bc.CurrentPlan()
	if err != nil {
		return false, err
	}
	if plan != nil {
		fmt.Printf("%sCurrent Upgrade Plan:\n", SPACE)
		fmt.Printf("%s%sName: %s\n", SPACE, SPACE, plan.Name)
		fmt.Printf("%s%sHeight: %d\n", SPACE, SPACE, plan.Height)
	}
	if len(chainId) > 0 && chainId != genesisId {
		fmt.Printf("%s%sWarning: The remote node chainId(%s) does not match the local genesis chainId(%s)\n", SPACE, SPACE, chainId, genesisId)
		return false, nil
	}
	if chainId == fxtypes.MainnetChainId {
		if blockHeight < fxtypes.MainnetBlockHeightV2 {
			fmt.Printf("%sVersion: V1\n", SPACE)
		} else if blockHeight < fxtypes.MainnetBlockHeightV3 {
			fmt.Printf("%sVersion: V2\n", SPACE)
		} else if blockHeight < fxtypes.MainnetBlockHeightV4 {
			fmt.Printf("%sVersion: v3\n", SPACE)
		} else if blockHeight < fxtypes.MainnetBlockHeightV5 {
			fmt.Printf("%sVersion: V4\n", SPACE)
		} else if blockHeight < fxtypes.MainnetBlockHeightV6 {
			fmt.Printf("%sVersion: V5\n", SPACE)
		} else if blockHeight < fxtypes.MainnetBlockHeightV7 {
			fmt.Printf("%sVersion: V6\n", SPACE)
		}
	}
	if chainId == fxtypes.TestnetChainId {
		if blockHeight < fxtypes.TestnetBlockHeightV2 {
			fmt.Printf("%sVersion: V1\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV3 {
			fmt.Printf("%sVersion: V2\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV4 {
			fmt.Printf("%sVersion: V3\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV41 {
			fmt.Printf("%sVersion: V4\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV42 {
			fmt.Printf("%sVersion: V4.1\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV5 {
			fmt.Printf("%sVersion: V4.2\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV6 {
			fmt.Printf("%sVersion: V5.0\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV7 {
			fmt.Printf("%sVersion: V6.0\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV71 {
			fmt.Printf("%sVersion: V7.0\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV72 {
			fmt.Printf("%sVersion: V7.1\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV73 {
			fmt.Printf("%sVersion: V7.2\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV74 {
			fmt.Printf("%sVersion: V7.3\n", SPACE)
		} else if blockHeight < fxtypes.TestnetBlockHeightV75 {
			fmt.Printf("%sVersion: V7.4\n", SPACE)
		}
	}
	return plan != nil, nil
}

func checkAppConfig(viper *viper.Viper) error {
	fmt.Printf("App Config:\n")
	fmt.Printf("%sFile: %s\n", SPACE, viper.ConfigFileUsed())
	config.SetConfigTemplate(fxcfg.DefaultConfigTemplate())
	appConfig := fxcfg.DefaultConfig()
	if err := viper.Unmarshal(appConfig); err != nil {
		return err
	}
	if err := appConfig.ValidateBasic(); err != nil {
		fmt.Printf("%sWarning: %s\n", SPACE, err.Error())
	}
	if appConfig.TLS.KeyPath == "" && appConfig.TLS.CertificatePath != "" {
		fmt.Printf("%sWarning: certificate_path is set, but key_path is not set\n", SPACE)
	}
	if appConfig.TLS.KeyPath != "" && appConfig.TLS.CertificatePath == "" {
		fmt.Printf("%sWarning: key_path is set, but certificate_path is not set\n", SPACE)
	}
	if appConfig.IAVLCacheSize > 781250 {
		fmt.Printf("%sWarning: if the node does not require high API query, you can appropriately reduce iavl_cache_size to reduce memory usage\n", SPACE)
	}
	return nil
}

func checkTmConfig(config *tmcfg.Config, needUpgrade bool) error {
	fmt.Printf("Tendermint Config:\n")
	fmt.Printf("%sFile: %s/config/config.toml\n", SPACE, config.RootDir)
	if err := config.ValidateBasic(); err != nil {
		fmt.Printf("%sWarning: ", err.Error())
	}
	if needUpgrade && config.Consensus.DoubleSignCheckHeight > 0 {
		fmt.Printf("%sWarning: double_sign_check_height is greater than 0\n", SPACE)
		fmt.Printf("%s%sPlease check the upgrade plan and set double_sign_check_height to 0\n", SPACE, SPACE)
		fmt.Printf("%s%sIf you are sure that the upgrade has been completed, you can ignore this warning\n", SPACE, SPACE)
	}
	if config.P2P.Seeds == "" {
		fmt.Printf("%sWarning: seeds is empty\n", SPACE)
	}
	if config.RPC.TLSKeyFile == "" && config.RPC.TLSCertFile != "" {
		fmt.Printf("%sWarning: tls_cert_file is not empty, but tls_key_file is empty\n", SPACE)
	}
	if config.RPC.TLSKeyFile != "" && config.RPC.TLSCertFile == "" {
		fmt.Printf("%sWarning: tls_key_file is not empty, but tls_cert_file is empty\n", SPACE)
	}
	return nil
}

//gocyclo:ignore
func checkCosmovisor(rootPath string, bc blockchain) error {
	cosmovisorPath := filepath.Join(rootPath, "cosmovisor")
	if _, err := os.Stat(cosmovisorPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Cosmovisor: %s\n", "not found")
			return nil
		}
		return err
	}
	defer func() {
		_ = printDirectory(cosmovisorPath, 0, []bool{false}, SPACE)
	}()

	fmt.Printf("Cosmovisor:\n")
	for _, dir := range []string{"current", "genesis"} {
		fmt.Printf("%s%s%s:\n", SPACE, strings.ToUpper(dir[:1]), dir[1:])
		fxcored := filepath.Join(cosmovisorPath, dir, "bin/fxcored")
		fmt.Printf("%s%sBinary File: %s\n", SPACE, SPACE, fxcored)
		output, err := exec.Command(fxcored, "version").Output()
		if err != nil {
			fmt.Printf("%s%sWarning: %s\n", SPACE, SPACE, err.Error())
			return nil
		}
		v := string(bytes.Trim(output, "\n"))
		fmt.Printf("%s%sfxcored version: %s\n", SPACE, SPACE, v)
	}

	upgradesPath := filepath.Join(cosmovisorPath, "upgrades")
	fmt.Printf("%sUpgrades:\n", SPACE)
	fmt.Printf("%s%sPath: %s\n", SPACE, SPACE, upgradesPath)
	entries, err := os.ReadDir(upgradesPath)
	if err != nil {
		fmt.Printf("%s%sWarning: %s\n", SPACE, SPACE, err.Error())
		return nil
	}
	plan, _ := bc.CurrentPlan()
	var planVersion bool
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if plan != nil && plan.Name == entry.Name() {
			planVersion = true
		}
		fmt.Printf("%s%s%s:\n", SPACE, SPACE, entry.Name())
		fxcored := filepath.Join(upgradesPath, entry.Name(), "bin/fxcored")
		fmt.Printf("%s%s%sBinary File: %s\n", SPACE, SPACE, SPACE, fxcored)
		output, err := exec.Command(fxcored, "version").Output()
		if err != nil {
			fmt.Printf("%s%s%sWarning: %s\n", SPACE, SPACE, SPACE, err.Error())
			continue
		}
		outputVersion := string(bytes.Trim(output, "\n"))
		fmt.Printf("%s%s%sfxcored version: %s\n", SPACE, SPACE, SPACE, outputVersion)

		if checkVersionCompatibility(entry.Name(), outputVersion) {
			fmt.Printf("%s%s%sWarning: fxcored version is not match upgrade plan\n", SPACE, SPACE, SPACE)
		}

		upgradeInfoFile := filepath.Join(upgradesPath, entry.Name(), upgradetypes.UpgradeInfoFilename)
		upgradeInfo, err := os.ReadFile(upgradeInfoFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("%s%s%sWarning: %s\n", SPACE, SPACE, SPACE, err.Error())
			}
			continue
		}
		fmt.Printf("%s%s%sUpgrade Info File: %s\n", SPACE, SPACE, SPACE, upgradeInfoFile)
		var plan upgradetypes.Plan
		if err := json.Unmarshal(upgradeInfo, &plan); err != nil {
			fmt.Printf("%s%s%sWarning: %s\n", SPACE, SPACE, SPACE, err.Error())
			continue
		}
		fmt.Printf("%s%s%sUpgrade Plan: %s %d\n", SPACE, SPACE, SPACE, plan.Name, plan.Height)
		if plan.Name != entry.Name() {
			fmt.Printf("%s%s%sWarning: fxcored version is not match upgrade plan\n", SPACE, SPACE, SPACE)
		}
	}
	if plan != nil && !planVersion {
		fmt.Printf("%s%sWarning: current upgrade plan is not found in cosmovisor\n", SPACE, SPACE)
	}
	return nil
}

func checkVersionCompatibility(version, outputVersion string) bool {
	semVer := strings.Split(version, ".")
	isFxV := len(semVer) == 1 && strings.HasPrefix(semVer[0], "fx") && strings.Contains(outputVersion, semVer[0][2:])
	isSemVer := len(semVer) > 1 && strings.Contains(outputVersion, semVer[0]+"."+semVer[1])
	return !isFxV && !isSemVer
}

func printDirectory(path string, depth int, last []bool, tab string) error {
	printPath := path
	if depth > 0 {
		printPath = filepath.Base(path)
	}
	printTree(printPath, depth, last, tab)

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for idx, entry := range entries {
		currentLast := idx == len(entries)-1
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if (info.Mode() & os.ModeSymlink) == os.ModeSymlink {
			fullPath, err := os.Readlink(filepath.Join(path, entry.Name()))
			if err != nil {
				return err
			} else {
				printTree(entry.Name()+" -> "+fullPath, depth+1, append(last, currentLast), tab)
			}
		} else if entry.IsDir() {
			if err = printDirectory(filepath.Join(path, entry.Name()), depth+1, append(last, currentLast), tab); err != nil {
				return err
			}
		} else {
			printTree(entry.Name(), depth+1, append(last, currentLast), tab)
		}
	}
	return nil
}

func printTree(entry string, depth int, last []bool, tab string) {
	if depth == 0 {
		fmt.Printf("%s%s\n", tab, entry)
	} else {
		indent := ""
		newLast := last[1:]
		for i := 0; i < len(newLast)-1; i++ {
			if newLast[i] {
				indent = fmt.Sprintf("%s    ", indent)
			} else {
				indent = fmt.Sprintf("%s│   ", indent)
			}
		}
		sepStr := "├── "
		if last[len(last)-1] {
			sepStr = "└── "
		}
		fmt.Printf("%s%s%s%s\n", tab, indent, sepStr, entry)
	}
}

func getGenesisDocAndSha256(genesisFile string) (*tmtypes.GenesisDoc, string, error) {
	genesisFileData, err := os.ReadFile(genesisFile)
	if err != nil {
		return nil, "", err
	}
	dst := &bytes.Buffer{}
	if err = json.Compact(dst, genesisFileData); err != nil {
		return nil, "", err
	}
	genesisDoc, err := tmtypes.GenesisDocFromJSON(genesisFileData)
	if err != nil {
		return nil, "", err
	}
	return genesisDoc, fxtypes.Sha256Hex(dst.Bytes()), nil
}
