package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	client := http.Client{Timeout: time.Duration(1)*time.Second}

	for {
		resp, err := client.Get("http://localhost:8080/limited")
		if err != nil {
			log.Printf("Error %s\n", err)
		}
		body, err := io.ReadAll(resp.Body)
		log.Printf("Response body: %s\n", body)
		time.Sleep(time.Duration(3)*time.Second)
	}
}