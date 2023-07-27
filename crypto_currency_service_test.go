package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

func TestGetCryptoCurrencyByID(t *testing.T) {
	// Create a new mock database and expected result
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cryptoService := NewCryptoCurrencyService(db)

	cryptoID := 1
	expectedCrypto := CryptoCurrency{
		ID:         cryptoID,
		Name:       "Bitcoin",
		UpVote:     100,
		DownVote:   20,
		TotalVotes: 120,
	}

	// Set the expectations for the first QueryRow call (count query)
	rowsCount := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crypto_vote WHERE id = ?").
		WithArgs(cryptoID).
		WillReturnRows(rowsCount)

	// Set the expectations for the second QueryRow call (get cryptocurrency query)
	rowsCrypto := sqlmock.NewRows([]string{"id", "name", "up_vote", "down_vote", "total_votes"}).
		AddRow(expectedCrypto.ID, expectedCrypto.Name, expectedCrypto.UpVote, expectedCrypto.DownVote, expectedCrypto.TotalVotes)
	mock.ExpectQuery("SELECT id, name, up_vote, down_vote, \\(up_vote \\+ down_vote\\) as total_votes FROM crypto_vote WHERE id=?").
		WithArgs(cryptoID).
		WillReturnRows(rowsCrypto)

	// Create a new request and recorder for testing the handler
	req, err := http.NewRequest("GET", "/v1/cryptovote/"+strconv.Itoa(cryptoID), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Set up the router and call the handler
	r := mux.NewRouter()
	r.HandleFunc("/v1/cryptovote/{id:[0-9]+}", cryptoService.GetCryptoCurrencyByID).Methods("GET")
	r.ServeHTTP(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse the response body into a CryptoCurrency object
	var actualCrypto CryptoCurrency
	err = json.Unmarshal(rr.Body.Bytes(), &actualCrypto)
	assert.NoError(t, err)

	// Check the response content
	assert.Equal(t, expectedCrypto, actualCrypto)

	// Check that all the expected SQL queries were executed
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCryptoCurrency(t *testing.T) {
	// Create a new mock database and expected result
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cryptoService := NewCryptoCurrencyService(db)

	// Mock the database query to check if the cryptocurrency name already exists
	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crypto_vote WHERE name = ?").
		WithArgs("Bitcoin").
		WillReturnRows(rows)

	// Mock the database insert to create a new cryptocurrency
	result := sqlmock.NewResult(1, 1) // Last insert ID: 1, Rows affected: 1
	mock.ExpectExec("INSERT INTO crypto_vote \\(name\\) VALUES \\(\\?\\)").
		WithArgs("Bitcoin").
		WillReturnResult(result)

	// Create a new request and recorder for testing the handler
	jsonData := `{"name": "Bitcoin"}`
	req, err := http.NewRequest("POST", "/v1/cryptovote", strings.NewReader(jsonData))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// Set up the router and call the handler
	r := mux.NewRouter()
	r.HandleFunc("/v1/cryptovote", cryptoService.CreateCryptoCurrency).Methods("POST")
	r.ServeHTTP(rr, req)

	// Check the response status code
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Parse the response body into a CryptoCurrency object
	var createdCrypto CryptoCurrency
	err = json.Unmarshal(rr.Body.Bytes(), &createdCrypto)
	assert.NoError(t, err)

	// Check the response content
	expectedCrypto := CryptoCurrency{
		ID:     1,         // Last insert ID
		Name:   "Bitcoin",
		UpVote: 0,         // Default value for a new cryptocurrency
		DownVote: 0,       // Default value for a new cryptocurrency
		TotalVotes: 0,     // Default value for a new cryptocurrency
	}
	assert.Equal(t, expectedCrypto, createdCrypto)

	// Check that all the expected SQL queries were executed
	assert.NoError(t, mock.ExpectationsWereMet())
}

