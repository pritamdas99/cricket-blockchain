package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const baseURL = "http://localhost:8080"

// 🔹 Replace with your actual connection IDs
var (
	cons = make(map[string]string)
)

func main() {
	GetConnectionID()
	createMatches()
	issueTickets()
	useTickets()
	distributeRevenue()
}

func post(url string, payload map[string]interface{}) {
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response:", url, resp.Status)
}

func createMatches() {
	file, _ := os.Open("../dataset/matches.csv")
	defer file.Close()

	reader := csv.NewReader(file)
	rows, _ := reader.ReadAll()

	fmt.Println("Creating Matches...")

	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}

		matchID := row[0]
		team1 := row[1]
		team2 := row[2]

		post(baseURL+"/create-match", map[string]interface{}{
			"connection_id": cons["ICCMSP"],
			"match_id":      matchID,
			"team1":         team1+"MSP",
			"team2":         team2+"MSP",
			"score1":        row[3],
			"score2":        row[4],
		})
	}
}

func issueTickets() {
	file, _ := os.Open("../dataset/tickets.csv")
	defer file.Close()

	reader := csv.NewReader(file)
	rows, _ := reader.ReadAll()

	fmt.Println("Issuing Tickets...")

	for i, row := range rows {
		if i == 0 {
			continue
		}

		ticketID := row[0]
		matchID := row[1]

		url := "http://localhost:8080/read"

		// 🔹 Request body
		payload := map[string]string{
			"match_id": matchID,
		}

		body, _ := json.Marshal(payload)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		// 🔹 Decode response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			fmt.Println("Decode error:", err)
			return
		}

		// 🔥 Extract data
		data, ok := result["data"].(map[string]interface{})
		if !ok {
			fmt.Println("No data found")
			return
		}

		fmt.Println("Match Data:", data, data["homeBoard"])

		post(baseURL+"/issue-ticket", map[string]interface{}{
			"connection_id": cons[data["homeBoard"].(string)],
			"match_id":      matchID,
			"ticket_id":     ticketID,
		})
	}
}

func useTickets() {
	file, _ := os.Open("../dataset/tickets.csv")
	defer file.Close()

	reader := csv.NewReader(file)
	rows, _ := reader.ReadAll()

	fmt.Println("Using Tickets...")

	for i, row := range rows {
		if i == 0 {
			continue
		}

		ticketID := row[0]
		matchID := row[1]

		url := "http://localhost:8080/read"

		// 🔹 Request body
		payload := map[string]string{
			"match_id": matchID,
		}

		body, _ := json.Marshal(payload)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		// 🔹 Decode response
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			fmt.Println("Decode error:", err)
			return
		}

		// 🔥 Extract data
		data, ok := result["data"].(map[string]interface{})
		if !ok {
			fmt.Println("No data found")
			return
		}

		fmt.Println("Match Data:", data)

		post(baseURL+"/use-ticket", map[string]interface{}{
			"connection_id": cons[data["homeBoard"].(string)],
			"ticket_id":     ticketID,
		})
	}
}

func distributeRevenue() {
	file, _ := os.Open("../dataset/revenue.csv")
	defer file.Close()

	reader := csv.NewReader(file)
	rows, _ := reader.ReadAll()

	fmt.Println("Distributing Revenue...")

	for i, row := range rows {
		if i == 0 {
			continue
		}

		matchID := row[0]
		fmt.Println(matchID)

		post(baseURL+"/revenue", map[string]interface{}{
			"connection_id": cons["ICCMSP"],
			"match_id":      matchID,
		})
	}
}

func GetConnectionID() {
	resp, err := http.Get("http://localhost:5021/connections")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	results := data["results"].([]interface{})

	for _, r := range results {
		conn := r.(map[string]interface{})

		label := conn["their_label"].(string)
		label = label+"MSP"
		connectionID := conn["connection_id"].(string)

		cons[label]=connectionID

	}

	fmt.Println(cons)

}