package main

import (
	//"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGetAllCryptoCurrencies(t *testing.T) {
	// Create a new mock database and expected result
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cryptoService := NewCryptoCurrencyService(db)

	rows := sqlmock.NewRows([]string{"id", "name", "up_vote", "down_vote", "total_votes"}).
		AddRow(1, "Bitcoin", 100, 20, 120).
		AddRow(2, "Ethereum", 80, 10, 90)

	mock.ExpectQuery("SELECT id, name, up_vote, down_vote, \\(up_vote \\+ down_vote\\) as total_votes FROM crypto_vote").
		WillReturnRows(rows)

	// Create a new request and recorder for testing the handler
	req, err := http.NewRequest("GET", "/v1/cryptovote", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Set up the router and call the handler
	r := mux.NewRouter()
	r.HandleFunc("/v1/cryptovote", cryptoService.GetAllCryptoCurrencies).Methods("GET")
	r.ServeHTTP(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse the response body into a slice of CryptoCurrency
	var cryptoCurrencies []CryptoCurrency
	err = json.Unmarshal(rr.Body.Bytes(), &cryptoCurrencies)
	assert.NoError(t, err)

	// Check the response content
	expectedCryptoCurrencies := []CryptoCurrency{
		{ID: 1, Name: "Bitcoin", UpVote: 100, DownVote: 20, TotalVotes: 120},
		{ID: 2, Name: "Ethereum", UpVote: 80, DownVote: 10, TotalVotes: 90},
	}
	assert.Equal(t, expectedCryptoCurrencies, cryptoCurrencies)

	// Check that all the expected SQL queries were executed
	assert.NoError(t, mock.ExpectationsWereMet())
}
