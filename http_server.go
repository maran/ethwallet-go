package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/eth-go"
	"log"
	"net/http"
)

// Ok so go doesn't do caps-locked constances, TODO: Change this
const ResponseOk = 1
const ResponseFailed = 0

const ErrorJson = 0
const ErrorForm = 1
const Error404 = 2

// Standard JSON error responses
type ErrorResponse struct {
	Status    int
	ErrorText string
	ErrorCode int
}

// Wrapper to create clear JSON error responses
func createJsonErrorResponse(w http.ResponseWriter, r *http.Request, httpCode int, errorCode int, errorText string) {
	e := ErrorResponse{Status: ResponseFailed, ErrorCode: errorCode, ErrorText: errorText}
	res, err := json.Marshal(e)
	if err != nil {
		// This should never happen
		panic("Creating json error response failed, help")
	}
	log.Println("Responding with JSON error: ", string(res))
	errorHandler(w, r, httpCode, string(res))
}

// Http error handler
func errorHandler(w http.ResponseWriter, r *http.Request, status int, data string) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprintf(w, "404:", data)
	} else if status == http.StatusInternalServerError {
		fmt.Fprintf(w, data)
	} else {
		fmt.Fprintf(w, data)
	}
}
func NewApiServer(addr string, stop chan bool, server *eth.Ethereum) {
	http.HandleFunc("/api/transactions", TransactionHandler)
	http.HandleFunc("/api/configs", ConfigHandler)

	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(htmlFolder()))))

	log.Println("Started API server, please visit http://localhost:4334")
	err := http.ListenAndServe(":4334", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
