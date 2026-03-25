package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

const (
	channelName   = "cricketchannel"
	chaincodeName = "sportscc"
	mspID         = "ICCMSP"
)

func main() {
	generateDataset()


}

func generateDataset() {
	boards := []string{"Board1", "Board2", "Board3"}

	totalTournaments := 2
	matchesPerTournament := 7
	ticketsPerMatch := "2"
	ticketPrice := "50"

	// -------- MATCH FILE --------
	matchFile, _ := os.Create("matches.csv")
	defer matchFile.Close()
	matchWriter := csv.NewWriter(matchFile)
	defer matchWriter.Flush()

	matchWriter.Write([]string{
		"id", "homeBoard", "awayBoard", "totalTickets", "ticketPrice",
	})

	// -------- TICKET FILE --------
	ticketFile, _ := os.Create("tickets.csv")
	defer ticketFile.Close()
	ticketWriter := csv.NewWriter(ticketFile)
	defer ticketWriter.Flush()

	ticketWriter.Write([]string{"code", "matchId"})

	// -------- REVENUE FILE --------
	revenueFile, _ := os.Create("revenue.csv")
	defer revenueFile.Close()
	revenueWriter := csv.NewWriter(revenueFile)
	defer revenueWriter.Flush()

	revenueWriter.Write([]string{"matchId"})

	for t := 1; t <= totalTournaments; t++ {

		for m := 1; m <= matchesPerTournament; m++ {

			matchID := fmt.Sprintf("tournament_%d_amatch_%d", t, m)

			home := boards[(m-1)%len(boards)]
			away := boards[m%len(boards)]

			// -------- MATCH --------
			matchWriter.Write([]string{
				matchID,
				home,
				away,
				fmt.Sprintf("%s", ticketsPerMatch),
				fmt.Sprintf("%s", ticketPrice),
			})

			// -------- TICKETS --------
			total, _ := strconv.Atoi(ticketsPerMatch)
			for tk := 1; tk <= total; tk++ {
				ticketWriter.Write([]string{
					fmt.Sprintf("%s_ticket_%d", matchID, tk),
					matchID,
				})
			}

			// -------- REVENUE --------
			revenueWriter.Write([]string{
				matchID,
			})
		}
	}
}