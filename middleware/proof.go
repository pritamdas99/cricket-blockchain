package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	// "fmt"
	"net/http"
)

func SendProofRequest(connectionID string) string {

	body := map[string]interface{}{
		"connection_id": connectionID,
		"auto_remove": false,
        "auto_verify": true,
		"presentation_request": map[string]interface{}{
			"indy": map[string]interface{}{
				"name": "Board Verification",
				"version": "1.0",
				"requested_attributes": map[string]interface{}{
					"attr1_referent": map[string]string{"name": "board_name"},
					"attr2_referent": map[string]string{"name": "role"},
				},
				"requested_predicates": map[string]interface{}{},
			},
		},
	}

	jsonBody, _ := json.Marshal(body)

	resp, _ := http.Post(
		"http://localhost:5021/present-proof-2.0/send-request",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Println("here it is")

	// fmt.Println(result)

	return result["pres_ex_id"].(string)
}