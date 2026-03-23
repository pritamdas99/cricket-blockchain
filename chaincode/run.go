package main

import (
	"context"
	"crypto/x509"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	channelName   = "cricketchannel"
	chaincodeName = "sportscc"
	mspID         = "ICCMSP"
)

func main() {

	// =============================
	// Generate synthetic dataset
	// =============================
	generateDataset()

	// =============================
	// Setup connection
	// =============================
	certPath := "crypto/users/Admin@icc.example.com/msp/signcerts/cert.pem"
	keyPath := "crypto/users/Admin@icc.example.com/msp/keystore/key.pem"
	tlsCertPath := "crypto/peers/peer0.icc.example.com/tls/ca.crt"

	cert, err := loadCertificate(certPath)
	if err != nil {
		log.Fatal(err)
	}

	id := identity.NewX509Identity(mspID, cert)

	sign, err := identity.NewPrivateKeySign(keyPath)
	if err != nil {
		log.Fatal(err)
	}

	certificate, err := os.ReadFile(tlsCertPath)
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certificate)

	creds := credentials.NewClientTLSFromCert(certPool, "")
	conn, err := grpc.Dial("localhost:7051", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(conn),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer gateway.Close()

	network := gateway.GetNetwork(channelName)
	contract := network.GetContract(chaincodeName)

	_, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// =============================
	// Invoke createMatch
	// =============================
	file, _ := os.Open("matches.csv")
	reader := csv.NewReader(file)
	reader.Read() // skip header

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		matchID := record[0]
		venue := record[1]
		date := record[2]

		fmt.Println("Submitting createMatch:", matchID)
		_, err = contract.SubmitTransaction("createMatch", matchID, venue, date)
		if err != nil {
			log.Fatal(err)
		}
	}

	file.Close()

	// =============================
	// Invoke issueTicket
	// =============================
	file2, _ := os.Open("tickets.csv")
	reader2 := csv.NewReader(file2)
	reader2.Read()

	for {
		record, err := reader2.Read()
		if err == io.EOF {
			break
		}

		ticketID := record[0]
		matchID := record[1]
		seat := record[2]
		price := record[3]

		fmt.Println("Submitting issueTicket:", ticketID)
		_, err = contract.SubmitTransaction("issueTicket", ticketID, matchID, seat, price)
		if err != nil {
			log.Fatal(err)
		}
	}

	file2.Close()

	fmt.Println("All transactions submitted successfully.")
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return identity.CertificateFromPEM(certificatePEM)
}

func generateDataset() {

	matchFile, _ := os.Create("matches.csv")
	matchWriter := csv.NewWriter(matchFile)
	matchWriter.Write([]string{"matchId", "venue", "date"})

	for i := 1; i <= 3; i++ {
		matchWriter.Write([]string{
			fmt.Sprintf("MATCH_%d", i),
			fmt.Sprintf("Stadium_%d", i),
			fmt.Sprintf("2026-03-0%d", i),
		})
	}
	matchWriter.Flush()
	matchFile.Close()

	ticketFile, _ := os.Create("tickets.csv")
	ticketWriter := csv.NewWriter(ticketFile)
	ticketWriter.Write([]string{"ticketId", "matchId", "seat", "price"})

	for i := 1; i <= 3; i++ {
		for j := 1; j <= 5; j++ {
			ticketWriter.Write([]string{
				fmt.Sprintf("TICKET_%d_%d", i, j),
				fmt.Sprintf("MATCH_%d", i),
				fmt.Sprintf("%d", j),
				"50",
			})
		}
	}
	ticketWriter.Flush()
	ticketFile.Close()
}
