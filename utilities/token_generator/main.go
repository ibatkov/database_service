package main

import (
	"database-service/auth"
	"flag"
	"fmt"
	"log"
	"time"
)

func main() {

	// Please replace the line above with actual import path for auth package.
	userId := flag.Int("userId", 100, "User Id to generate JWT token for")
	phrase := flag.String("phrase", "example-phrase", "Secret phrase of JWT token")
	flag.Parse()
	authService := auth.NewJwtService(*phrase) //
	authService.TokenTTL = 24 * time.Hour
	if token, err := authService.GenerateToken(*userId); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(token)
	}
}
