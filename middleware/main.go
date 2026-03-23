package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	

	fmt.Println("starting")

	InitFabric()

	http.HandleFunc("/create-match", CreateMatchHandler)
	http.HandleFunc("/issue-ticket", IssueTicketHandler)
	http.HandleFunc("/use-ticket", UseTicketHandler)
	http.HandleFunc("/revenue", RevenueHandler)
	http.HandleFunc("/read", Read)

	log.Println("🚀 Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}