package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// ===============================
// ASSETS
// ===============================

type Match struct {
	ID           string  `json:"id"`
	HomeBoard    string  `json:"homeBoard"`
	AwayBoard    string  `json:"awayBoard"`
	TotalTickets int     `json:"totalTickets"`
	TicketsSold  int     `json:"ticketsSold"`
	TicketPrice  float64 `json:"ticketPrice"`
	RevenuePool  float64 `json:"revenuePool"`
	Status       string  `json:"status"` // SCHEDULED, COMPLETED
}

type Ticket struct {
	Code     string `json:"code"`
	MatchID  string `json:"matchId"`
	OwnerID  string `json:"ownerId"`
	IssuedBy string `json:"issuedBy"`
	Used     bool   `json:"used"`
}

type BoardAccount struct {
	BoardID string  `json:"boardId"`
	Balance float64 `json:"balance"`
}

type RevenueRecord struct {
	ID          string  `json:"id"`
	MatchID     string  `json:"matchId"`
	TotalAmount float64 `json:"totalAmount"`
	HomeShare   float64 `json:"homeShare"`
	AwayShare   float64 `json:"awayShare"`
	ICCShare    float64 `json:"iccShare"`
	Settled     bool    `json:"settled"`
}

// ===============================
// MATCH CREATION (ICC ONLY)
// ===============================

func (s *SmartContract) CreateMatch(ctx contractapi.TransactionContextInterface,
	id string,
	home string,
	away string,
	totalTickets string,
	ticketPrice string,
) error {

	mspID, _ := ctx.GetClientIdentity().GetMSPID()
	if mspID != "ICCMSP" {
		return fmt.Errorf("only ICC can create match")
	}

	tt, _ := strconv.Atoi(totalTickets)
	price, _ := strconv.ParseFloat(ticketPrice, 64)

	match := Match{
		ID:           id,
		HomeBoard:    home,
		AwayBoard:    away,
		TotalTickets: tt,
		TicketsSold:  0,
		TicketPrice:  price,
		RevenuePool:  0,
		Status:       "SCHEDULED",
	}

	bytes, _ := json.Marshal(match)
	return ctx.GetStub().PutState("MATCH_"+id, bytes)
}

// ===============================
// ISSUE TICKET (After Payment)
// ===============================

func (s *SmartContract) IssueTicket(ctx contractapi.TransactionContextInterface,
	matchID string,
	ticketCode string,
) error {

	clientID, _ := ctx.GetClientIdentity().GetID()
	mspID, _ := ctx.GetClientIdentity().GetMSPID()

	matchBytes, _ := ctx.GetStub().GetState("MATCH_" + matchID)
	if matchBytes == nil {
		return fmt.Errorf("match not found")
	}

	var match Match
	json.Unmarshal(matchBytes, &match)

	if match.HomeBoard != mspID {
		return fmt.Errorf("not the board that hosting match")
	}

	if match.TicketsSold >= match.TotalTickets {
		return fmt.Errorf("sold out")
	}

	existing, _ := ctx.GetStub().GetState("TICKET_" + ticketCode)
	if existing != nil {
		return fmt.Errorf("ticket already exists")
	}

	// Increase revenue pool
	match.TicketsSold++
	match.RevenuePool += match.TicketPrice

	ticket := Ticket{
		Code:     ticketCode,
		MatchID:  matchID,
		OwnerID:  clientID,
		IssuedBy: mspID,
		Used:     false,
	}

	matchJSON, _ := json.Marshal(match)
	ticketJSON, _ := json.Marshal(ticket)

	ctx.GetStub().PutState("MATCH_"+matchID, matchJSON)
	return ctx.GetStub().PutState("TICKET_"+ticketCode, ticketJSON)
}

// ===============================
// USE TICKET
// ===============================

func (s *SmartContract) UseTicket(ctx contractapi.TransactionContextInterface,
	ticketCode string,
) error {

	clientID, _ := ctx.GetClientIdentity().GetID()

	bytes, _ := ctx.GetStub().GetState("TICKET_" + ticketCode)
	if bytes == nil {
		return fmt.Errorf("ticket not found %v", ticketCode)
	}

	var ticket Ticket
	json.Unmarshal(bytes, &ticket)

	if ticket.OwnerID != clientID {
		return fmt.Errorf("not owner")
	}

	if ticket.Used {
		return fmt.Errorf("already used")
	}

	ticket.Used = true

	newBytes, _ := json.Marshal(ticket)
	return ctx.GetStub().PutState("TICKET_"+ticketCode, newBytes)
}

// ===============================
// REVENUE DISTRIBUTION (ICC ONLY)
// ===============================

func (s *SmartContract) DistributeRevenue(ctx contractapi.TransactionContextInterface,
	matchID string,
) error {

	mspID, _ := ctx.GetClientIdentity().GetMSPID()
	if mspID != "ICCMSP" {
		return fmt.Errorf("only ICC can distribute revenue")
	}

	matchBytes, _ := ctx.GetStub().GetState("MATCH_" + matchID)
	if matchBytes == nil {
		return fmt.Errorf("match not found")
	}

	var match Match
	json.Unmarshal(matchBytes, &match)

	total := match.RevenuePool

	homeShare := total * 0.4
	awayShare := total * 0.4
	iccShare := total * 0.2

	// Update board balances
	s.updateBoard(ctx, match.HomeBoard, homeShare)
	s.updateBoard(ctx, match.AwayBoard, awayShare)
	s.updateBoard(ctx, "ICC", iccShare)

	record := RevenueRecord{
		ID:          "REV_" + matchID,
		MatchID:     matchID,
		TotalAmount: total,
		HomeShare:   homeShare,
		AwayShare:   awayShare,
		ICCShare:    iccShare,
		Settled:     true,
	}

	recordBytes, _ := json.Marshal(record)

	// Reset revenue pool
	match.RevenuePool = 0
	match.Status = "COMPLETED"


	ctx.GetStub().DelState("MATCH_"+matchID)

	return ctx.GetStub().PutState("REV_"+matchID, recordBytes)
}

// Helper
func (s *SmartContract) updateBoard(ctx contractapi.TransactionContextInterface,
	board string,
	amount float64,
) error {

	key := "BOARD_" + board
	bytes, _ := ctx.GetStub().GetState(key)

	var account BoardAccount

	if bytes == nil {
		account = BoardAccount{BoardID: board, Balance: amount}
	} else {
		json.Unmarshal(bytes, &account)
		account.Balance += amount
	}

	newBytes, _ := json.Marshal(account)
	return ctx.GetStub().PutState(key, newBytes)
}

// ===============================
// QUERY FUNCTIONS
// ===============================

func (s *SmartContract) ReadMatch(ctx contractapi.TransactionContextInterface,
	id string,
) (*Match, error) {

	bytes, _ := ctx.GetStub().GetState("MATCH_" + id)
	if bytes == nil {
		return nil, fmt.Errorf("not found")
	}

	var match Match
	json.Unmarshal(bytes, &match)
	return &match, nil
}

func main() {
	cc, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
