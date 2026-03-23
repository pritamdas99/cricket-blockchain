package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func WaitForProof(presExID string) (bool, string) {

	for {

		resp, err := http.Get("http://localhost:5021/present-proof-2.0/records/" + presExID)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		state := result["state"].(string)
		fmt.Println("⏳ State:", state)

		if state == "done" {

			attrs := result["by_format"].(map[string]interface{})["pres"].(map[string]interface{})["indy"].(map[string]interface{})["requested_proof"].(map[string]interface{})["self_attested_attrs"].(map[string]interface{})

			role := attrs["attr1_referent"].(string)

			fmt.Println("role ", role)

			return true, role
		}

		time.Sleep(2 * time.Second)
	}
}