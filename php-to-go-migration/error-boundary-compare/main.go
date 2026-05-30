package main

import (
	"encoding/json"
	"fmt"
)

type UserPayload struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

func main() {
	input := []byte(`{"user_id":"42","email":123}`)

	var payload UserPayload
	if err := json.Unmarshal(input, &payload); err != nil {
		fmt.Printf("decode error: %T: %v\n", err, err)
		return
	}

	fmt.Printf("decoded: user_id=%d email=%s\n", payload.UserID, payload.Email)
}
