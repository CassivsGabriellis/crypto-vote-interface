package main

import (
	"fmt"
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
		ID:         1, // Last insert ID
		Name:       "Bitcoin",
		UpVote:     0, // Default value for a new cryptocurrency
		DownVote:   0, // Default value for a new cryptocurrency
		TotalVotes: 0, // Default value for a new cryptocurrency
	}
	assert.Equal(t, expectedCrypto, createdCrypto)

	// Check that all the expected SQL queries were executed
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestVoteCryptoCurrency(t *testing.T) {
	// Common setup for both upvote and downvote tests
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cryptoService := NewCryptoCurrencyService(db)
	cryptoID := 1

	// Set the expectations for the first QueryRow call (count query)
	rowsCount := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crypto_vote WHERE id = ?").
		WithArgs(cryptoID).
		WillReturnRows(rowsCount)

	// Set the expectations for the second Exec call (update votes query)
	result := sqlmock.NewResult(1, 1) // Rows affected: 1
	mock.ExpectExec("UPDATE crypto_vote SET .* WHERE id = ?").
		WithArgs(cryptoID).
		WillReturnResult(result)

	// Set the expectations for the third QueryRow call (get updated cryptocurrency query)
	rowsCrypto := sqlmock.NewRows([]string{"id", "name", "up_vote", "down_vote", "total_votes"}).
		AddRow(1, "Bitcoin", 100, 21, 121)
	mock.ExpectQuery("SELECT id, name, up_vote, down_vote, \\(up_vote \\+ down_vote\\) as total_votes FROM crypto_vote WHERE id=?").
		WithArgs(cryptoID).
		WillReturnRows(rowsCrypto)

	t.Run("UpVote", func(t *testing.T) {
		// Create a new request and recorder for upvote testing
		req, err := http.NewRequest("PUT", "/v1/cryptovote/"+strconv.Itoa(cryptoID)+"/upvote", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()

		// Set up the router and call the handler for upvote
		r := mux.NewRouter()
		r.HandleFunc("/v1/cryptovote/{id:[0-9]+}/upvote", cryptoService.UpVoteCryptoCurrency).Methods("PUT")
		r.ServeHTTP(rr, req)

		// Check the response status code
		assert.Equal(t, http.StatusOK, rr.Code)

		// Parse the response body into a CryptoCurrency object
		var updatedCrypto CryptoCurrency
		err = json.Unmarshal(rr.Body.Bytes(), &updatedCrypto)
		assert.NoError(t, err)

		// Check the response content for upvote
		expectedCrypto := CryptoCurrency{
			ID:         1,
			Name:       "Bitcoin",
			UpVote:     100,
			DownVote:   21,
			TotalVotes: 121,
		}
		assert.Equal(t, expectedCrypto, updatedCrypto)

		// Check that all the expected SQL queries were executed for upvote
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DownVote", func(t *testing.T) {
		// Set the expectations for the first QueryRow call (count query for downvote)
		rowsCountDownVote := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crypto_vote WHERE id = ?").
			WithArgs(cryptoID).
			WillReturnRows(rowsCountDownVote)
	
		// Set the expectations for the second Exec call (update votes query for downvote)
		resultDownVote := sqlmock.NewResult(1, 1) // Rows affected: 1
		mock.ExpectExec("UPDATE crypto_vote SET .* WHERE id = ?").
			WithArgs(cryptoID).
			WillReturnResult(resultDownVote)
	
		// Set the expectations for the third QueryRow call (get updated cryptocurrency query for downvote)
		rowsCryptoDownVote := sqlmock.NewRows([]string{"id", "name", "up_vote", "down_vote", "total_votes"}).
			AddRow(1, "Bitcoin", 100, 21, 121) // Corrected values for downvote
		mock.ExpectQuery("SELECT id, name, up_vote, down_vote, \\(up_vote \\+ down_vote\\) as total_votes FROM crypto_vote WHERE id=?").
			WithArgs(cryptoID).
			WillReturnRows(rowsCryptoDownVote)
	
		// Create a new request and recorder for downvote testing
		req, err := http.NewRequest("PUT", "/v1/cryptovote/"+strconv.Itoa(cryptoID)+"/downvote", nil)
		assert.NoError(t, err)
	
		rr := httptest.NewRecorder()
	
		// Set up the router and call the handler for downvote
		r := mux.NewRouter()
		r.HandleFunc("/v1/cryptovote/{id:[0-9]+}/downvote", cryptoService.DownVoteCryptoCurrency).Methods("PUT")
		r.ServeHTTP(rr, req)
	
		// Check the response status code
		assert.Equal(t, http.StatusOK, rr.Code)
	
		// Parse the response body into a CryptoCurrency object
		var updatedCrypto CryptoCurrency
		err = json.Unmarshal(rr.Body.Bytes(), &updatedCrypto)
		assert.NoError(t, err)
	
		// Check the response content for downvote
		expectedCryptoDownVote := CryptoCurrency{
			ID:         1,
			Name:       "Bitcoin",
			UpVote:     100,
			DownVote:   21,
			TotalVotes: 121,
		}
		assert.Equal(t, expectedCryptoDownVote, updatedCrypto)
	
		// Check that all the expected SQL queries were executed for downvote
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteCryptoCurrency(t *testing.T) {
	t.Run("DeleteExistingCrypto", func(t *testing.T) {
		// Create a new mock database and expected result
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		cryptoService := NewCryptoCurrencyService(db)
		cryptoID := 1

		// Set the expectations for the first QueryRow call (count query)
		rowsCount := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crypto_vote WHERE id = ?").
			WithArgs(cryptoID).
			WillReturnRows(rowsCount)

		// Set the expectations for the Exec call (delete cryptocurrency query)
		result := sqlmock.NewResult(1, 1) // Rows affected: 1
		mock.ExpectExec("DELETE FROM crypto_vote WHERE id = ?").
			WithArgs(cryptoID).
			WillReturnResult(result)

		// Create a new request and recorder for testing the handler
		req, err := http.NewRequest("DELETE", "/v1/cryptovote/"+strconv.Itoa(cryptoID), nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()

		// Set up the router and call the handler
		r := mux.NewRouter()
		r.HandleFunc("/v1/cryptovote/{id:[0-9]+}", cryptoService.DeleteCryptoCurrency).Methods("DELETE")
		r.ServeHTTP(rr, req)

		// Check the response status code
		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Check that all the expected SQL queries were executed
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteNonExistingCrypto", func(t *testing.T) {
		// Create a new mock database and expected result
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		cryptoService := NewCryptoCurrencyService(db)
		cryptoID := 1

		// Set the expectations for the first QueryRow call (count query for non-existing cryptocurrency)
		rowsCount := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crypto_vote WHERE id = ?").
			WithArgs(cryptoID).
			WillReturnRows(rowsCount)

		// Create a new request and recorder for testing the handler
		req, err := http.NewRequest("DELETE", "/v1/cryptovote/"+strconv.Itoa(cryptoID), nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()

		// Set up the router and call the handler
		r := mux.NewRouter()
		r.HandleFunc("/v1/cryptovote/{id:[0-9]+}", cryptoService.DeleteCryptoCurrency).Methods("DELETE")
		r.ServeHTTP(rr, req)

		// Check the response status code
		assert.Equal(t, http.StatusNotFound, rr.Code)

		// Check that all the expected SQL queries were executed
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseErrorOnCountQuery", func(t *testing.T) {
		// Create a new mock database and simulate an error during the count query
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		cryptoService := NewCryptoCurrencyService(db)
		cryptoID := 1

		// Set the expectations for the first QueryRow call (count query)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crypto_vote WHERE id = ?").
			WithArgs(cryptoID).
			WillReturnError(fmt.Errorf("database error"))

		// Create a new request and recorder for testing the handler
		req, err := http.NewRequest("DELETE", "/v1/cryptovote/"+strconv.Itoa(cryptoID), nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()

		// Set up the router and call the handler
		r := mux.NewRouter()
		r.HandleFunc("/v1/cryptovote/{id:[0-9]+}", cryptoService.DeleteCryptoCurrency).Methods("DELETE")
		r.ServeHTTP(rr, req)

		// Check the response status code
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		// Check that all the expected SQL queries were executed
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseErrorOnDeleteQuery", func(t *testing.T) {
		// Create a new mock database and simulate an error during the delete query
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		cryptoService := NewCryptoCurrencyService(db)
		cryptoID := 1

		// Set the expectations for the first QueryRow call (count query)
		rowsCount := sqlmock.NewRows([]string{"count"}).AddRow(1)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM crypto_vote WHERE id = ?").
			WithArgs(cryptoID).
			WillReturnRows(rowsCount)

		// Set the expectations for the Exec call (delete cryptocurrency query with error)
		mock.ExpectExec("DELETE FROM crypto_vote WHERE id = ?").
			WithArgs(cryptoID).
			WillReturnError(fmt.Errorf("database error"))

		// Create a new request and recorder for testing the handler
		req, err := http.NewRequest("DELETE", "/v1/cryptovote/"+strconv.Itoa(cryptoID), nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()

		// Set up the router and call the handler
		r := mux.NewRouter()
		r.HandleFunc("/v1/cryptovote/{id:[0-9]+}", cryptoService.DeleteCryptoCurrency).Methods("DELETE")
		r.ServeHTTP(rr, req)

		// Check the response status code
		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		// Check that all the expected SQL queries were executed
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
