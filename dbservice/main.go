package main

import (
	api "database-service/dbservice/api"
)

func main() {
	dbServiceApi, err := api.NewApi()
	if err != nil {
		return
	}
	dbServiceApi.Run()
}
