package keys

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

const (
	flagUnarmoredHex = "unarmored-hex"
	flagUnsafe       = "unsafe"
	flagASCIIArmored = "ascii-armor"
)

// ExportKeyCommand exports private keys from the key store.
func ExportKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <name>",
		Short: "Export private keys",
		Long: `Export a private key from the local keyring in encrypted format.

When both the --unarmored-hex and --unsafe flags are selected, cryptographic
private key material is exported in an INSECURE fashion that is designed to
allow users to import their keys in hot wallets. This feature is for advanced
users only that are confident about how to handle private keys work and are
FULLY AWARE OF THE RISKS. If you are unsure, you may want to do some research
and export your keys in encrypted format.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			buf := bufio.NewReader(clientCtx.Input)
			unarmored, _ := cmd.Flags().GetBool(flagUnarmoredHex)
			unsafe, _ := cmd.Flags().GetBool(flagUnsafe)

			if unarmored && unsafe {
				return exportUnsafeUnarmored(cmd, args[0], buf, clientCtx.Keyring)
			} else if unarmored || unsafe {
				return fmt.Errorf("the flags %s and %s must be used together", flagUnsafe, flagUnarmoredHex)
			}

			encryptPassword, err := input.GetPassword("Enter passphrase to encrypt the exported key:", buf)
			if err != nil {
				return err
			}

			asciiArmored, _ := cmd.Flags().GetBool(flagASCIIArmored)
			if asciiArmored {
				armored, err := clientCtx.Keyring.ExportPrivKeyArmor(args[0], encryptPassword)
				if err != nil {
					return err
				}

				cmd.Println(armored)

				return nil
			}

			hexPrivKey, err := keyring.NewUnsafe(clientCtx.Keyring).UnsafeExportPrivKeyHex(args[0])
			if err != nil {
				return err
			}
			priv, err := ethcrypto.HexToECDSA(hexPrivKey)
			if err != nil {
				return err
			}
			key := &keystore.Key{
				PrivateKey: priv,
				Address:    ethcrypto.PubkeyToAddress(priv.PublicKey),
			}
			keyjson, err := keystore.EncryptKey(key, encryptPassword, keystore.StandardScryptN, keystore.StandardScryptP)
			if err != nil {
				return err
			}

			cmd.Println(string(keyjson))

			return nil
		},
	}

	cmd.Flags().Bool(flagUnarmoredHex, false, "Export unarmored hex privkey. Requires --unsafe.")
	cmd.Flags().Bool(flagUnsafe, false, "Enable unsafe operations. This flag must be switched on along with all unsafe operation-specific options.")
	cmd.Flags().Bool(flagASCIIArmored, false, "Enable ASCII-armored encrypted format")
	return cmd
}

func exportUnsafeUnarmored(cmd *cobra.Command, uid string, buf *bufio.Reader, kr keyring.Keyring) error {
	// confirm deletion, unless -y is passed
	if yes, err := input.GetConfirmation("WARNING: The private key will be exported as an unarmored hexadecimal string. USE AT YOUR OWN RISK. Continue?", buf, cmd.ErrOrStderr()); err != nil {
		return err
	} else if !yes {
		return nil
	}

	hexPrivKey, err := keyring.NewUnsafe(kr).UnsafeExportPrivKeyHex(uid)
	if err != nil {
		return err
	}

	cmd.Print(hexPrivKey)

	return nil
}

func ImportKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <name> <keyfile>",
		Short: "Import private keys into the local keybase",
		Long:  "Import a ASCII armored or ethereum keystore or unencrypted private key into the local keybase.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			buf := bufio.NewReader(clientCtx.Input)

			bz, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}
			//ioutil.ReadFile read all data, contains line break at the end of a file
			bz = bytes.TrimSuffix(bz, []byte{'\n'})

			if len(bz) == 64 {
				priv, err := ethcrypto.HexToECDSA(string(bz))
				if err != nil {
					return err
				}
				armoPrivKey := crypto.EncryptArmorPrivKey(&ethsecp256k1.PrivKey{Key: ethcrypto.FromECDSA(priv)}, "", ethsecp256k1.KeyType)
				return clientCtx.Keyring.ImportPrivKey(args[0], armoPrivKey, "")
			}
			passphrase, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
			if err != nil {
				return err
			}
			key, err := keystore.DecryptKey(bz, passphrase)
			if err == nil {
				armoPrivKey := crypto.EncryptArmorPrivKey(&ethsecp256k1.PrivKey{Key: ethcrypto.FromECDSA(key.PrivateKey)}, "", ethsecp256k1.KeyType)
				return clientCtx.Keyring.ImportPrivKey(args[0], armoPrivKey, "")
			}

			return clientCtx.Keyring.ImportPrivKey(args[0], string(bz), passphrase)
		},
	}
	return cmd
}
