package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var API_ORIGIN = "https://ecodadys.app"

func main() {
	fmt.Println("----------------------")
	fmt.Println("- ECODADYS DOWNLOADER -")
	fmt.Println("-----------------------")
	fmt.Println()

	user, err := login()
	if err != nil {
		panic(err)
	}

	fmt.Println(user)
}

type User struct {
	id    float64
	token string
}

func login() (User, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please login into your account.")

	fmt.Print("Enter your email: ")
	scanner.Scan()
	email := scanner.Text()

	fmt.Print("Enter your password (default: ecodadys): ")
	scanner.Scan()
	password := scanner.Text()
	if password == "" {
		password = "ecodadys"
	}

	fmt.Println("Logging in...")

	body := map[string]any{
		"device_type": map[string]any{
			"string": "android",
			"valid":  true,
		},
		"username": email,
		"password": password,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return User{}, fmt.Errorf("error encoding JSON: %v", err)
	}

	endpoint := API_ORIGIN + "/api/api/user/login"
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return User{}, err
	}
	req.Header.Add("content-type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return User{}, fmt.Errorf("error decoding response: %v", err)
	}

	id, ok := result["id"].(float64)
	if !ok {
		return User{}, fmt.Errorf("user id not found")
	}

	tokenData, ok := result["token"].(map[string]any)
	if !ok {
		return User{}, fmt.Errorf("token data not found")
	}

	token, ok := tokenData["string"].(string)
	if !ok {
		return User{}, fmt.Errorf("token string not found")
	}

	fmt.Println("Successfully logged in")
	return User{
		id:    id,
		token: token,
	}, nil
}
