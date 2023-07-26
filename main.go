package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/CassivsGabriellis/crypto-vote-interface/handler"
)

func main() {
	// Set up for the routes
	myRouter := mux.NewRouter()

	myRouter.HandleFunc("/cryptocurrencies", handler.GetAllCryptoCurrencies).Methods("GET")
	myRouter.HandleFunc("/cryptocurrencies/{id}", handler.GetCryptoCurrencyByID).Methods("GET")
	myRouter.HandleFunc("/cryptocurrencies", handler.CreateCryptoCurrency).Methods("POST")
	myRouter.HandleFunc("/cryptocurrencies/{id}/upvote", handler.UpVoteCryptoCurrency).Methods("PUT")
	myRouter.HandleFunc("/cryptocurrencies/{id}/downvote", handler.DownVoteCryptoCurrency).Methods("PUT")
	myRouter.HandleFunc("/cryptocurrencies/{id}", handler.DeleteCryptoCurrency).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", myRouter))

	fmt.Println("Connection success!")
}
