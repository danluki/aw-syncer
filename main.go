package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type UploadResponse struct {
	OperationID string `json:"operation_id"`
	Href        string `json:"href"`
	Method      string `json:"method"`
	Templated   bool   `json:"templated"`
}

func main() {
	token := flag.String("token", "", "Yandex OAuth token")
	awServerAddr := flag.String("server", "http://localhost:5600", "ActivityWatch server address")
	flag.Parse()

	if len(*token) == 0 {
		panic("Token is required")
	}

	if awServerAddr == nil {
		panic("ActivityWatch server is required")
	}

	r, err := http.Get(fmt.Sprintf("%s/api/0/export", *awServerAddr))
	if err != nil {
		panic(fmt.Sprintf("Failed to get data from ActivityWatch: %+v", err))
	}
	defer r.Body.Close()

	currentDateTime := time.Now().Format("2006-01-02")

	fileName := fmt.Sprintf("%s.json", currentDateTime)
	file, err := os.Create(fileName)
	if err != nil {
		panic(fmt.Sprintf("Failed to create file with ActivityWatch data: %+v", err))
	}
	defer file.Close()

	_, err = io.Copy(file, r.Body)
	if err != nil {
		panic(fmt.Sprintf("Failed to save file with ActivityWatch data: %+v", err))
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		panic(fmt.Sprintf("Failed to rewind file cursor: %+v", err))
	}

	fmt.Printf("Saved ActivityWatch data to %s\n", fileName)

	//Send data to disk
	yurl := "https://cloud-api.yandex.net/v1/disk/resources/upload/?path=activitywatch&overwrite=true"

	request, err := http.NewRequest("GET", yurl, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Set the Authorization header with the OAuth token
	request.Header.Set("Authorization", "OAuth "+*token)

	// Send the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("Failed to use Yandex api: %+v", err))
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		panic(fmt.Sprintf("Failed to load upload body: %+v", err))
	}

	var uploadResponse UploadResponse
	if err := json.Unmarshal(body, &uploadResponse); err != nil {
		panic(fmt.Sprintf("Failed to parse response: %+v", err))
	}

	uploadURL := uploadResponse.Href
	request, err = http.NewRequest("PUT", uploadURL, file)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse response: %+v", err))
	}
	request.Header.Set("Authorization", "OAuth "+*token)

	client = &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("Faield to upload file to Yandex disk: %+v", err))
	}
	defer resp.Body.Close()

	fmt.Println("Syncing done successfully.")
}
