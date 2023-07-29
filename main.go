package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
)

// CORS middleware implementation
func enableCorsMiddleware(next http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(next)
}

func main() {
	// Set up for the routes
	myRouter := mux.NewRouter()

	// Apply CORS middleware
	myRouter.Use(enableCorsMiddleware)

	// Add API version prefix to the routes
	apiRouter := myRouter.PathPrefix("/v1").Subrouter()

	// Initialize database
	db, err := initializeDB()
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer db.Close()

	// Initialize CryptoCurrency service with database connection
	cryptoService := NewCryptoCurrencyService(db)

	// Register API endpoints with handlers
	apiRouter.HandleFunc("/cryptovote", cryptoService.GetAllCryptoCurrencies).Methods("GET")
	apiRouter.HandleFunc("/cryptovote/{id:[0-9]+}", cryptoService.GetCryptoCurrencyByID).Methods("GET")
	apiRouter.HandleFunc("/cryptovote", cryptoService.CreateCryptoCurrency).Methods("POST")
	apiRouter.HandleFunc("/cryptovote/{id:[0-9]+}/upvote", cryptoService.UpVoteCryptoCurrency).Methods("PUT")
	apiRouter.HandleFunc("/cryptovote/{id:[0-9]+}/downvote", cryptoService.DownVoteCryptoCurrency).Methods("PUT")
	apiRouter.HandleFunc("/cryptovote/{id:[0-9]+}", cryptoService.DeleteCryptoCurrency).Methods("DELETE")

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not set in environment variable
	}

	serverAddress := fmt.Sprintf(":%s", port)
	log.Fatal(http.ListenAndServe(serverAddress, myRouter))
}