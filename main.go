package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Set up for the routes
	myRouter := mux.NewRouter()

	myRouter.HandleFunc("/cryptocurrencies", GetAllCryptoCurrencies).Methods("GET")
	myRouter.HandleFunc("/cryptocurrencies/{id}", GetCryptoCurrencyByID).Methods("GET")
	myRouter.HandleFunc("/cryptocurrencies", CreateCryptoCurrency).Methods("POST")
	myRouter.HandleFunc("/cryptocurrencies/{id}/upvote", UpVoteCryptoCurrency).Methods("PUT")
	myRouter.HandleFunc("/cryptocurrencies/{id}/downvote", DownVoteCryptoCurrency).Methods("PUT")
	myRouter.HandleFunc("/cryptocurrencies/{id}", DeleteCryptoCurrency).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", myRouter))

	fmt.Println("Connection success!")
}
