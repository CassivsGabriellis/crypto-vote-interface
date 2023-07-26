package main

type CryptoCurrency struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	UpVote     int    `json:"up_vote"`
	DownVote   int    `json:"down_vote"`
	TotalVotes int    `json:"total_votes"`
}
