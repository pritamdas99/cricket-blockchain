package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var contracts = map[string]*client.Contract{}

func InitFabric() {
	
	contracts["ICCMSP"] = connect("ICCMSP", 
	"../config/crypto/peerOrganizations/icc.example.com/users/Admin@icc.example.com/msp/signcerts/Admin@icc.example.com-cert.pem", 
	"../config/crypto/peerOrganizations/icc.example.com/users/Admin@icc.example.com/msp/keystore/priv_sk")
	contracts["Board1MSP"] = connect("Board1MSP", 
    "../config/crypto/peerOrganizations/board1.example.com/users/Admin@board1.example.com/msp/signcerts/Admin@board1.example.com-cert.pem", 
	"../config/crypto/peerOrganizations/board1.example.com/users/Admin@board1.example.com/msp/keystore/priv_sk")
	contracts["Board2MSP"] = connect("Board2MSP", 
    "../config/crypto/peerOrganizations/board2.example.com/users/Admin@board2.example.com/msp/signcerts/Admin@board2.example.com-cert.pem", 
	"../config/crypto/peerOrganizations/board2.example.com/users/Admin@board2.example.com/msp/keystore/priv_sk")

	fmt.Println("Fabric initialized", contracts["Board1MSP"])
}

func connect(mspID, certPath, keyPath string) *client.Contract {

	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		fmt.Println("readfile ", err)
		return nil
	}
	cert, err := loadCert(certPEM)
	if err != nil {

		fmt.Println("loadcert ",err)
		return nil
	}

	id, _ := identity.NewX509Identity(mspID, cert)

	fmt.Println("got the id stuff")

	privateKey, err := loadPrivateKey(keyPath)
	if err != nil {
		panic(err)
	}

	signer, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	fmt.Println("got the sign stuff")

	xx := "../config/crypto/peerOrganizations/%v.example.com/peers/peer0.%v.example.com/tls/ca.crt"

	tlsCert, _ := os.ReadFile(fmt.Sprintf(xx,"icc","icc"))
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(tlsCert)


	creds := credentials.NewTLS(&tls.Config{
		RootCAs: certPool,
		ServerName: "peer0.icc.example.com",
	})

	conn, _ := grpc.Dial("localhost:6051", grpc.WithTransportCredentials(creds))

	fmt.Println("got the conn stuff")

	gw, _ := client.Connect(
		id,
		client.WithSign(signer),
		client.WithClientConnection(conn),
	)

	network := gw.GetNetwork("cricketchannel")

	return network.GetContract("cricketcc")
}

func SubmitTxAs(role string, fn string, args ...string) {

	msp := roleToMSP(role)

	contract, ok := contracts[msp]
	if !ok {
		fmt.Println("❌ Unknown MSP:", msp)
		return
	}

	fmt.Println("➡️ Acting as:", msp)

	// 🔥 Use Submit() instead of SubmitTransaction()
	// res, err := contract.Submit(
	// 	fn,
	// 	client.WithArguments(args...),
	// 	client.WithEndorsingOrganizations("ICCMSP", "Board1MSP", "Board2MSP"),
	// )
	// if err != nil {
	// 	fmt.Println("❌ Fabric error:", err)
	// 	return
	// }

	// res, err := contract.SubmitTransaction(fn, args...)
	// if err != nil {
	// 	fmt.Println("Error: ",err)
	// 	return
	// }

	// fmt.Println("TxID:", string(res))

	result, commit, err := contract.SubmitAsync(
		fn,
		client.WithArguments(args...),
		client.WithEndorsingOrganizations("ICCMSP", "Board1MSP", "Board2MSP"),
	)
	if err != nil {
		fmt.Println("Err: ", err)
	}

	// ✅ Get TxID here
	txID := commit.TransactionID()
	fmt.Println("TxID:", txID)

	// Wait for commit (optional but recommended)
	status, err := commit.Status()
	if err != nil {
		panic(err)
	}

	

	fmt.Println("✅ Transaction success33", status.Successful, string(result))
}