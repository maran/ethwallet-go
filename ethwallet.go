package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/ethereum/eth-go"
	"github.com/ethereum/ethchain-go"
	"github.com/ethereum/ethutil-go"
	"log"
	"math/big"
	"os"
	"os/signal"
)

var createConfig = flag.Bool("create-config", false, "Create a new wallet")
var action = flag.String("action", "", "What do you wan tot do")
var to = flag.String("to", "me", "Recipient address")
var amount = flag.Int64("amount", 1, "Amount to send")

var Config WalletConfig
var EthServer *eth.Ethereum

// Retrieving balance: data := CurrentBlock.State().Get(ACCOUNT_BYTES)
// return struct Account := NewAddressFromData([]byte(data))
// Nonce zit in de struct van account

// Register interrupt handlers so we can stop the ethereum
func RegisterInterupts(s *eth.Ethereum) {
	// Buffered chan of one is enough
	c := make(chan os.Signal, 1)
	// Notify about interrupts for now
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("Shutting down (%v) ... \n", sig)

			s.Stop()
		}
	}()
}

func Init() {
	flag.Parse()

	if *createConfig {
		createConfigFile()
	}

	ethchain.InitFees()
	ethutil.ReadConfig()
}

func createAndBroadcastTx(recipient string, amount big.Int, password []byte) (*ethchain.Transaction, error) {
	data := []string{}

	tx := ethchain.NewTransaction([]byte(recipient), &amount, data)
	// We should be adding a new nonce based on saved transactions
	tx.Nonce = Config.GetServerAccount().Nonce

	privKey := Config.GetPrivateKey(password)

	// TODO: Regenerate the public key from this private key and compare it to the saved public key in the config file so we know the pass was correct.
	tx.Sign(privKey)

	// Sender() returns a public key which was used to sign so compare sender with option.publickey()
	sender := hex.EncodeToString(tx.Sender())

	if sender == Config.Account {
		hash := hex.EncodeToString(tx.Hash())
		fmt.Println("Transaction", hash, " created and broadcasted.")
		Config.DeserializedTxs = append(Config.DeserializedTxs, tx)

		EthServer.TxPool.QueueTransaction(tx)

		Config.LastBalance = Config.GetServerAccount().Amount.String()
		Config.Save()

		return tx, nil
	} else {
		return tx, errors.New("Wrong password")
	}

}

func handleTransaction() {
	transactionChannel := make(ethchain.TxPoolHook, 1)
	EthServer.TxPool.Hook = transactionChannel
	for {
		select {
		case tx := <-transactionChannel:
			sender := tx.Sender()
			receiver := tx.Recipient

			// We might want to handle these differently in the future
			if bytes.Equal(sender, Config.GetAccount()) {
				Config.Transactions = append(Config.Transactions, tx.RlpEncode())
			} else if bytes.Equal(receiver, Config.GetAccount()) {
				Config.Transactions = append(Config.Transactions, tx.RlpEncode())
			}
			// TODO: Why is this empty
			/*
				ff :=
					fmt.Println("CURRENT STATE: ", ff)
			*/

			Config.LastBalance = Config.GetServerAccount().Amount.String()
			Config.Save()

		}
	}
	// TODO: Set hook to nil when gracefully shutting down.
}

func main() {
	// Parse command flags
	Init()

	// Setup ethereum
	ethereum, err := eth.New(eth.CapDefault, true)
	if err != nil {
		log.Println(err)
		return
	}
	RegisterInterupts(ethereum)
	ethereum.Start()

	// Load config
	loadOrCreateConfig()

	EthServer = ethereum

	// Start transaction handling
	go handleTransaction()

	// network stats
	// len(EthServer.Peers)

	// Handle actions
	fmt.Println("Wallet loaded\nmain account:", Config.Account)

	Config.DecodeAndLoadTransactions()

	Config.LastBalance = Config.GetServerAccount().Amount.String()

	if *action == "create-tx" {
		fmt.Println("Creating new transaction")
		bigint := big.NewInt(*amount)
		_, e := createAndBroadcastTx(*to, *bigint, nil)
		if e != nil {
			fmt.Println(e)
		}
	}

	if _, err := os.Stat(htmlFolder()); os.IsNotExist(err) {
		fmt.Println("Ethereum wallet HTML not found, not starting webserver.")
		fmt.Println("If you want HTML support please clone the repository from github.com/pdisle/ethereum-wallet-html.")
	} else {
		NewApiServer("notimplementedyet", make(chan bool), ethereum)
	}

	ethereum.WaitForShutdown()
}
