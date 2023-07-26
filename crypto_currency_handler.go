package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func GetCryptoCurrencyByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	cryptoID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid cryptocurrency ID", http.StatusBadRequest)
		return
	}

	var crypto CryptoCurrency

	err = db.QueryRow("SELECT id, name, up_vote, down_vote, (up_vote + down_vote) as total_votes FROM crypto_vote WHERE id=?", cryptoID).
		Scan(&crypto.ID, &crypto.Name, &crypto.UpVote, &crypto.DownVote, &crypto.TotalVotes)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Error getting cryptocurrency", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(crypto)
}

func GetAllCryptoCurrencies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cryptoCurrencies := []CryptoCurrency{}

	rows, err := db.Query("SELECT id, name, up_vote, down_vote, (up_vote + down_vote) as total_votes FROM crypto_vote")
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Error getting cryptocurrencies", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var crypto CryptoCurrency
		if err := rows.Scan(&crypto.ID, &crypto.Name, &crypto.UpVote, &crypto.DownVote, &crypto.TotalVotes); err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Error getting cryptocurrencies", http.StatusInternalServerError)
			return
		}
		cryptoCurrencies = append(cryptoCurrencies, crypto)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error iterating over rows:", err)
		http.Error(w, "Error getting cryptocurrencies", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cryptoCurrencies)
}

func CreateCryptoCurrency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var crypto CryptoCurrency

	err := json.NewDecoder(r.Body).Decode(&crypto)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Perform additional validation
	if crypto.Name == "" {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}

	if _, err := strconv.Atoi(crypto.Name); err == nil {
		http.Error(w, "Name cannot be a number", http.StatusBadRequest)
		return
	}

	// Set votes to zero for a new cryptocurrency
	crypto.UpVote = 0
	crypto.DownVote = 0

	// Check if the cryptocurrency name already exists in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM crypto_vote WHERE name = ?", crypto.Name).Scan(&count)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Error creating cryptocurrency", http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "Cryptocurrency with this name already exists", http.StatusConflict)
		return
	}

	// Validation successful, insert into the database
	result, err := db.Exec("INSERT INTO crypto_vote (name) VALUES (?)",
		crypto.Name)
	if err != nil {
		log.Println("Error inserting into database:", err)
		http.Error(w, "Error creating cryptocurrency", http.StatusInternalServerError)
		return
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		log.Println("Error getting last insert ID:", err)
		http.Error(w, "Error creating cryptocurrency", http.StatusInternalServerError)
		return
	}

	crypto.ID = int(lastInsertID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(crypto)
}

func UpVoteCryptoCurrency(w http.ResponseWriter, r *http.Request) {
	voteCryptoCurrency(w, r, "up_vote")
}

func DownVoteCryptoCurrency(w http.ResponseWriter, r *http.Request) {
	voteCryptoCurrency(w, r, "down_vote")
}

func voteCryptoCurrency(w http.ResponseWriter, r *http.Request, voteType string) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	cryptoID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid cryptocurrency ID", http.StatusBadRequest)
		return
	}

	// Check if the cryptocurrency exists in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM crypto_vote WHERE id = ?", cryptoID).Scan(&count)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Error voting for cryptocurrency", http.StatusInternalServerError)
		return
	}

	if count == 0 {
		http.Error(w, "Cryptocurrency does not exist", http.StatusNotFound)
		return
	}

	// Perform the vote (increment or decrement by one)
	var voteColumn string
	switch voteType {
	case "up_vote":
		voteColumn = "up_vote"
	case "down_vote":
		voteColumn = "down_vote"
	default:
		http.Error(w, "Invalid vote type", http.StatusBadRequest)
		return
	}

	// Update the database with the new vote count
	_, err = db.Exec("UPDATE crypto_vote SET "+voteColumn+" = "+voteColumn+" + 1, total_votes = up_vote + down_vote WHERE id = ?", cryptoID)
	if err != nil {
		log.Println("Error updating database:", err)
		http.Error(w, "Error voting for cryptocurrency", http.StatusInternalServerError)
		return
	}

	// Retrieve the updated cryptocurrency
	var crypto CryptoCurrency
	err = db.QueryRow("SELECT id, name, up_vote, down_vote, (up_vote + down_vote) as total_votes FROM crypto_vote WHERE id=?", cryptoID).
		Scan(&crypto.ID, &crypto.Name, &crypto.UpVote, &crypto.DownVote, &crypto.TotalVotes)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Error getting cryptocurrency", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(crypto)
}

func DeleteCryptoCurrency(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	cryptoID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid cryptocurrency ID", http.StatusBadRequest)
		return
	}

	// Check if the cryptocurrency exists in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM crypto_vote WHERE id = ?", cryptoID).Scan(&count)
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Error deleting cryptocurrency", http.StatusInternalServerError)
		return
	}

	if count == 0 {
		http.Error(w, "Cryptocurrency does not exist", http.StatusNotFound)
		return
	}

	// Delete the cryptocurrency from the database
	_, err = db.Exec("DELETE FROM crypto_vote WHERE id = ?", cryptoID)
	if err != nil {
		log.Println("Error deleting from database:", err)
		http.Error(w, "Error deleting cryptocurrency", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
