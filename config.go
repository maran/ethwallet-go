package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/ethchain-go"
	"github.com/ethereum/ethutil-go"
	"github.com/obscuren/secp256k1-go"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
)

type WalletConfig struct {
	PublicKey       []byte
	EncodedPrivKey  []byte
	Account         string
	LastBalance     string
	Transactions    [][]byte
	DeserializedTxs []*ethchain.Transaction
}

func (config *WalletConfig) GetServerAccount() *ethchain.Address {
	return EthServer.BlockManager.BlockChain().CurrentBlock.GetAddr(config.GetAccount())

}

// Get a Byte array as account instead of string
func (config *WalletConfig) GetAccount() []byte {
	var ba []byte
	var err error

	ba, err = hex.DecodeString(config.Account)
	if err != nil {
		log.Panic(err)
	}
	return ba
}

func loadOrCreateConfig() {
	path := configFile()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Println("Ethereum wallet config file does not exist.")
		var e error
		Config, e = createConfigFile()
		if e != nil {
			log.Println("Error creating wallet file, exiting")
		}
	} else {
		log.Println("Ethereum wallet config exists, loading.")
		Config = loadConfigFile()
	}
}

func loadConfigFile() WalletConfig {
	var config WalletConfig

	fileContents, err := ioutil.ReadFile(configFile())
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(fileContents, &config)
	if err != nil {
		fmt.Println("Could not read config file, most likely corrupted.")
		os.Exit(1)
	}
	return config
}

func (config *WalletConfig) DecodeAndLoadTransactions() []*ethchain.Transaction {
	var txs []*ethchain.Transaction

	for _, tx := range config.Transactions {
		newTx := ethchain.NewTransactionFromData(tx)
		txs = append(txs, newTx)
	}
	config.DeserializedTxs = txs
	return txs
}

func (config *WalletConfig) GetPrivateKey(password []byte) []byte {
	if password == nil {
		password = getPassword()
	}

	privKey := decrypt(password, config.EncodedPrivKey)
	return privKey
}

func (config *WalletConfig) Save() bool {
	json, err := json.Marshal(config)
	if err != nil {
		fmt.Println("Couldn't save wallet")
	} else {
		err = ioutil.WriteFile(configFile(), json, 0644)
		if err != nil {
			panic(err)
		}
		log.Println("Wallet saved")
	}
	return true
}

func createConfigFile() (WalletConfig, error) {
	// Create folder
	err := os.MkdirAll(configFolder(), 0700)

	if err == nil {
		// And config file
		createIt := true

		// Double check the file does not exist already
		if _, err := os.Stat(configFile()); err == nil {
			fmt.Println("Warning: Wallet file already exists, are you sure you want to overwrite? (y/n)")
			var whatToDo string
			_, err := fmt.Scanf("%s", &whatToDo)
			if err != nil {
				createIt = false
			}
			if whatToDo != "yes" {
				createIt = false
			}
		}
		if createIt {
			_, e := os.Create(configFile())
			err = e
		} else {
			fmt.Println("Not creating wallet file")
			return WalletConfig{}, nil
		}
	}

	fmt.Println("Creating keypair")
	pub, priv := secp256k1.GenerateKeyPair()
	addr := hex.EncodeToString(ethutil.Sha3Bin(pub)[12:])
	password := getConfirmPassword()

	if len(password) < 8 {
		fmt.Println("Warning: Looks like you have chosen a pretty short password, you might want to increase the length.")
	}

	fmt.Println("Encrypting private keys")
	fmt.Println("Your account is", addr)

	encodedPrivKey := encrypt(password, priv)
	// Create address..?
	config := WalletConfig{PublicKey: pub, EncodedPrivKey: encodedPrivKey, Account: addr}

	config.Save()

	return config, err
}
func htmlFolder() string {
	return path.Join(configFolder(), "/ethereum-wallet-html")
}
func configFolder() string {
	usr, _ := user.Current()
	path := path.Join(usr.HomeDir, ".ethereum")
	return path
}

func configFile() string {
	return path.Join(configFolder(), "wallet.dat")
}
