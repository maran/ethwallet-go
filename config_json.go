package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Path[len("/api/configs"):]

	log.Println("Handling method:", r.Method, "with action:", action)

	switch r.Method {
	case "POST":
		switch action {
		case "": // Create new transaction
		default:
			createJsonErrorResponse(w, r, http.StatusNotFound, Error404, fmt.Sprint("No action: ", r.Method, action))
		}
	case "GET":
		switch action {
		case "":
			res, err := json.Marshal(Config)
			if err != nil {
				fmt.Println("Nope", err.Error())
			} else {
				fmt.Fprintf(w, string(res))
			}
		}
	default:
		createJsonErrorResponse(w, r, http.StatusNotFound, Error404, fmt.Sprint("No action: ", r.Method, action))

	}
}
