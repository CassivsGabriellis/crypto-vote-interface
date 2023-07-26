package model

type CryptoCurrency struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	UpVote     int    `json:"up_vote"`
	DownVote   int    `json:"down_vote"`
	TotalVotes int    `json:"total_votes"`
}

// Add any additional methods specific to the CryptoCurrency struct, if needed.
// For example, you can have a method to validate the CryptoCurrency fields before insertion.