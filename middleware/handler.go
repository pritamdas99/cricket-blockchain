package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

/* =========================
   Request Structs
========================= */

type CreateMatchRequest struct {
	ConnectionID string `json:"connection_id"`

	MatchID string `json:"match_id"`
	Team1   string `json:"team1"`
	Team2   string `json:"team2"`
	Score1  string `json:"score1"`
	Score2  string `json:"score2"`
}

type IssueTicketRequest struct {
	ConnectionID string `json:"connection_id"`

	MatchID  string `json:"match_id"`
	TicketID string `json:"ticket_id"`
}

type Revenue struct {
	ConnectionID string `json:"connection_id"`

	MatchID  string `json:"match_id"`
}

type UseTicketRequest struct {
	ConnectionID string `json:"connection_id"`

	TicketID string `json:"ticket_id"`
}

/* =========================
   Handlers
========================= */

func CreateMatchHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Step 1: Send proof request
	presExID := SendProofRequest(req.ConnectionID)
	fmt.Println("📩 Proof request sent:", presExID)

	// Step 2: Wait for proof
	ok, role := WaitForProof(presExID)
	if !ok {
		http.Error(w, "Verification failed", http.StatusUnauthorized)
		return
	}

	// Step 3: Role check
	if role != "ICC" {
		http.Error(w, "Only ICC allowed", http.StatusForbidden)
		return
	}

	// Step 4: Submit transaction
	SubmitTxAs(role,
		"CreateMatch",
		req.MatchID,
		req.Team1,
		req.Team2,
		req.Score1,
		req.Score2,
	)

	fmt.Println("🎉 Match created")

	// Response
	json.NewEncoder(w).Encode(map[string]string{
		"status": "match created",
	})
}

func IssueTicketHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req IssueTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Step 1: Send proof request
	presExID := SendProofRequest(req.ConnectionID)
	fmt.Println("📩 Proof request sent:", presExID)

	// Step 2: Wait for proof
	ok, role := WaitForProof(presExID)
	if !ok {
		http.Error(w, "Verification failed", http.StatusUnauthorized)
		return
	}

	// Step 3: Role check
	if role == "ICC" {
		http.Error(w, "Only boards can issue ticket", http.StatusForbidden)
		return
	}

	// Step 4: Submit transaction
	SubmitTxAs(role,
		"IssueTicket",
		req.MatchID,
		req.TicketID,
	)

	fmt.Println("🎫 Ticket issued")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ticket issued",
	})
}

func UseTicketHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UseTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Step 1: Send proof request
	presExID := SendProofRequest(req.ConnectionID)
	fmt.Println("📩 Proof request sent:", presExID)

	// Step 2: Wait for proof
	ok, role := WaitForProof(presExID)
	if !ok {
		http.Error(w, "Verification failed", http.StatusUnauthorized)
		return
	}

	// Step 3: Role check
	if role == "ICC" {
		http.Error(w, "Only boards can check ticket", http.StatusForbidden)
		return
	}

	// Step 4: Submit transaction
	SubmitTxAs(role,
		"UseTicket",
		req.TicketID,
	)

	fmt.Println("🎫 Ticket used")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ticket used",
	})
}

func RevenueHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Revenue
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Step 1: Send proof request
	presExID := SendProofRequest(req.ConnectionID)
	fmt.Println("📩 Proof request sent:", presExID)

	// Step 2: Wait for proof
	ok, role := WaitForProof(presExID)
	if !ok {
		http.Error(w, "Verification failed", http.StatusUnauthorized)
		return
	}

	// Step 3: Role check
	if role != "ICC" {
		http.Error(w, "Only ICC allowed", http.StatusForbidden)
		return
	}

	// Step 4: Submit transaction
	SubmitTxAs(role,
		"DistributeRevenue",
		req.MatchID,
		
	)

	fmt.Println("🎉 Revenue Distributed")

	// Response
	json.NewEncoder(w).Encode(map[string]string{
		"status": "revenue distributed",
	})
}

func Read(w http.ResponseWriter, r *http.Request) {

	var req Revenue
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return 
	}
	// Step 4: Submit transaction
	data := SubmitTxAs("ICC",
		"ReadMatch",
		req.MatchID,
		
	)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	fmt.Println("🎉 Read")

	// Response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "read",
		"data": result,
	})

}