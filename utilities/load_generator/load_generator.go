package main

import (
	"database-service/auth"
	"flag"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var rpsFlag = flag.Int("rps", 100, "Count of requests per seconds")
var hostFlag = flag.String("host", "http://localhost:8080", "Requested host")

func main() {

	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	authService := auth.NewJwtService("example-phrase")
	authService.TokenTTL = 10 * time.Minute
	var wg sync.WaitGroup
	wg.Add(1)
	for _ = range *rpsFlag {
		go func() {
			httpClient := &http.Client{
				Transport: &http.Transport{
					IdleConnTimeout:    5 * time.Second,
					DisableCompression: true,
				},
			}
			ticker := time.NewTicker(1 * time.Second)
			for _ = range ticker.C {
				randomUserId := rand.Intn(500000) + 1
				token, err := authService.GenerateToken(randomUserId)
				if err != nil {
					return
				}
				req, err := http.NewRequest("GET", *hostFlag+"/data", nil)
				if err != nil {
					log.Fatalf("failed to create request: %v", err)
					return
				}

				// Add the authorization header to the req
				req.Header.Add("Authorization", "Bearer "+token)

				// Do the request
				_, err = httpClient.Do(req)
				if err != nil {
					log.Fatalf("failed to send request: %v", err)
					return
				}
				httpClient.CloseIdleConnections()
			}
		}()

	}
	wg.Wait()
}
