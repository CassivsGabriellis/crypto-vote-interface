package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupTestDB() (*sql.DB, sqlmock.Sqlmock) {
	// Create a new mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		panic("Error creating mock database")
	}

	// Set up expectations for the GetAllCryptoCurrencies query
	rows := sqlmock.NewRows([]string{"id", "name", "up_vote", "down_vote", "total_votes"}).
		AddRow(1, "Crypto1", 10, 5, 15).
		AddRow(2, "Crypto2", 20, 3, 23)

	mock.ExpectQuery("^SELECT id, name, up_vote, down_vote, \\(up_vote \\+ down_vote\\) as total_votes FROM crypto_vote$").
		WillReturnRows(rows)

	// Set up expectations for the GetCryptoCurrencyByID query with ID=1
	cryptoRows := sqlmock.NewRows([]string{"id", "name", "up_vote", "down_vote", "total_votes"}).
		AddRow(1, "Crypto1", 10, 5, 15)

	mock.ExpectQuery("^SELECT id, name, up_vote, down_vote, \\(up_vote \\+ down_vote\\) as total_votes FROM crypto_vote WHERE id=\\?$").
		WithArgs(1).
		WillReturnRows(cryptoRows)

	// Set up expectations for the GetCryptoCurrencyByID query with non-existing ID=999
	mock.ExpectQuery("^SELECT id, name, up_vote, down_vote, \\(up_vote \\+ down_vote\\) as total_votes FROM crypto_vote WHERE id=\\?$").
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	return db, mock
}

func TestGetAllCryptoCurrencies(t *testing.T) {
	// Set up the mock database
	db, mock := setupTestDB()
	defer db.Close()

	// Create a new test instance of CryptoCurrencyService with the mock database
	cryptoService := NewCryptoCurrencyService(db)

	// Create a new HTTP request to the test server
	req, err := http.NewRequest("GET", "/v1/cryptovote", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the GetAllCryptoCurrencies handler function
	handler := http.HandlerFunc(cryptoService.GetAllCryptoCurrencies)
	handler.ServeHTTP(rr, req)

	// Check the response status code (expecting 200 OK)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response content type (expecting application/json)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Parse the response body to a slice of CryptoCurrency
	var cryptoCurrencies []CryptoCurrency
	err = json.NewDecoder(rr.Body).Decode(&cryptoCurrencies)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Assert that the correct number of cryptocurrencies is returned (2 in this case)
	assert.Len(t, cryptoCurrencies, 2)

	// Assert the values of the first cryptocurrency
	assert.Equal(t, 1, cryptoCurrencies[0].ID)
	assert.Equal(t, "Crypto1", cryptoCurrencies[0].Name)
	assert.Equal(t, 10, cryptoCurrencies[0].UpVote)
	assert.Equal(t, 5, cryptoCurrencies[0].DownVote)
	assert.Equal(t, 15, cryptoCurrencies[0].TotalVotes)

	// Assert the values of the second cryptocurrency
	assert.Equal(t, 2, cryptoCurrencies[1].ID)
	assert.Equal(t, "Crypto2", cryptoCurrencies[1].Name)
	assert.Equal(t, 20, cryptoCurrencies[1].UpVote)
	assert.Equal(t, 3, cryptoCurrencies[1].DownVote)
	assert.Equal(t, 23, cryptoCurrencies[1].TotalVotes)

	// Ensure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Some expectations were not met: %v", err)
	}
}

func TestGetCryptoCurrencyByID(t *testing.T) {
	// Set up the mock database
	db, _ := setupTestDB()
	defer db.Close()

	// Create a new test instance of CryptoCurrencyService with the mock database
	cryptoService := NewCryptoCurrencyService(db)

	// Test scenario: Get existing crypto currency by ID
	req, err := http.NewRequest("GET", "/v1/cryptovote/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(cryptoService.GetCryptoCurrencyByID)
	handler.ServeHTTP(rr, req)

	// Check the response status code (expecting 200 OK)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse the response body to a CryptoCurrency instance
	var crypto CryptoCurrency
	err = json.NewDecoder(rr.Body).Decode(&crypto)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Assert the values of the retrieved cryptocurrency
	assert.Equal(t, 1, crypto.ID)
	assert.Equal(t, "Crypto1", crypto.Name)
	assert.Equal(t, 10, crypto.UpVote)
	assert.Equal(t, 5, crypto.DownVote)
	assert.Equal(t, 15, crypto.TotalVotes)

	// Test scenario: Get non-existing crypto currency by invalid ID
	req, err = http.NewRequest("GET", "/v1/cryptovote/999", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(cryptoService.GetCryptoCurrencyByID)
	handler.ServeHTTP(rr, req)

	// Check the response status code (expecting 404 Not Found)
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Test scenario: Get crypto currency by malformed ID
	req, err = http.NewRequest("GET", "/v1/cryptovote/invalid_id", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(cryptoService.GetCryptoCurrencyByID)
	handler.ServeHTTP(rr, req)

	// Check the response status code (expecting 400 Bad Request)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
