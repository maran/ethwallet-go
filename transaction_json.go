package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/ethchain-go"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
)

// Successful transaction JSON response
type TransactionResponse struct {
	Hash      string
	Amount    int64
	Status    int
	Recipient string
	Password  string
	ErrorText string
}

// Unified transaction object
type TransactionJson struct {
	Hash        string
	Recipient   string
	Sender      string
	Amount      *big.Int
	BlockHeight int
	BlockDate   int
}

func EncodeToFriendlyStruct(tx *ethchain.Transaction) TransactionJson {
	ftx := TransactionJson{
		Hash:      hex.EncodeToString(tx.Hash()),
		Recipient: hex.EncodeToString(tx.Recipient),
		Amount:    tx.Value,
		Sender:    hex.EncodeToString(tx.Sender()),
	}
	return ftx
}

// Handles transaction related queries
func TransactionHandler(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Path[len("/api/transactions"):]

	log.Println("Handling method", r.Method, "with action", action)

	switch r.Method {
	case "POST":
		switch action {
		case "": // Create new transaction
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				createJsonErrorResponse(w, r, http.StatusInternalServerError, ErrorForm, err.Error())
				return
			}
			var t TransactionResponse

			err = json.Unmarshal(body, &t)
			if err != nil {
				createJsonErrorResponse(w, r, http.StatusInternalServerError, ErrorJson, err.Error())
				return
			}
			pass := []byte(t.Password)

			tx, e := createAndBroadcastTx(t.Recipient, *big.NewInt(t.Amount), pass)

			var jsonResponse TransactionResponse
			if e == nil {
				jsonResponse = TransactionResponse{Amount: t.Amount, Recipient: t.Recipient, Status: ResponseOk, Hash: hex.EncodeToString(tx.Hash())}
			} else {
				jsonResponse = TransactionResponse{Amount: t.Amount, Recipient: t.Recipient, Status: ResponseFailed, ErrorText: e.Error()}
			}

			res, err := json.Marshal(jsonResponse)
			if err != nil {
				createJsonErrorResponse(w, r, http.StatusInternalServerError, ErrorJson, err.Error())
				return
			}
			fmt.Fprintf(w, string(res))
		default:
			createJsonErrorResponse(w, r, http.StatusNotFound, Error404, fmt.Sprint("No action: ", r.Method, action))
		}
	case "GET":
		switch action {
		case "":
			var txs []TransactionJson
			for _, tx := range Config.DeserializedTxs {
				txs = append(txs, EncodeToFriendlyStruct(tx))

			}

			if len(txs) == 0 {
				fmt.Fprintf(w, string("[]"))
			} else {

				res, err := json.Marshal(txs)
				if err != nil {
					fmt.Println("Nope", err.Error())
				} else {
					fmt.Fprintf(w, string(res))
				}
			}

		}
	default:
		createJsonErrorResponse(w, r, http.StatusNotFound, Error404, fmt.Sprint("No action: ", r.Method, action))

	}
}
