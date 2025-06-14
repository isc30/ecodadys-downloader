package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

	images, err := getResources(user, "images")
	if err != nil {
		panic(err)
	}

	videos, err := getResources(user, "videos")
	if err != nil {
		panic(err)
	}

	downloadResources(append(images, videos...))
}

type User struct {
	Id    float64
	Token string
}

func login() (User, error) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Please login into your account.")

	var email string
	for strings.TrimSpace(email) == "" {
		fmt.Print("Enter your email: ")
		scanner.Scan()
		email = scanner.Text()
	}

	fmt.Print("Enter your password (ecodadys): ")
	scanner.Scan()
	password := scanner.Text()
	if strings.TrimSpace(password) == "" {
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
		Id:    id,
		Token: token,
	}, nil
}

type Resource struct {
	Url string `json:"url"`
}

func getResources(user User, resourceType string) ([]string, error) {
	endpoint := API_ORIGIN + fmt.Sprintf("/api/api/multimedia_content/%v/%d", resourceType, int(user.Id))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+user.Token)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unreadable body: %v", err)
	}
	// fmt.Println(string(bodyBytes))

	var resources []Resource
	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&resources); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	var urls []string
	for _, r := range resources {
		urls = append(urls, r.Url)
	}
	return urls, nil
}

func downloadResources(resources []string) {
	var wg sync.WaitGroup
	folder := "downloads"

	for _, url := range resources {
		wg.Add(1)
		go downloadFromUrl(url, folder, &wg)
	}

	wg.Wait()
	fmt.Println("All downloads complete.")
}

func downloadFromUrl(url, folder string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Extract the file name from the URL
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	filePath := filepath.Join(folder, fileName)

	// Create the folder if it doesn't exist
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		fmt.Printf("Failed to create folder: %v\n", err)
		return
	}

	// Download the image
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to download %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Failed to create file %s: %v\n", filePath, err)
		return
	}
	defer out.Close()

	// Copy the content
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Failed to save %s: %v\n", filePath, err)
		return
	}

	fmt.Printf("Downloaded: %s\n", filePath)
}
