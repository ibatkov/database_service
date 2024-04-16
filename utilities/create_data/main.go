package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

var (
	usersFlag  = flag.Int("users", 1000, "Number of users to create")
	dataFlag   = flag.Int("data", 5, "Number of data to create for each user")
	workerFlag = flag.Int("workers", 6, "Number of workers")
)

const Data = `Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.`

func main() {
	flag.Parse()

	db, err := sql.Open("postgres", "host=localhost port=5432 user=database-service password=password dbname=users_data sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	tasks := make(chan int, *workerFlag)

	bar := progressbar.Default(int64(*usersFlag))

	var wg sync.WaitGroup
	for i := 0; i < *workerFlag; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range tasks {
				var userId int
				err := db.QueryRow(`INSERT INTO users (access_level) VALUES ($1) RETURNING id`, "user").Scan(&userId)
				if err != nil {
					log.Fatal(err)
				}

				dataQuery := `INSERT INTO data (user_id, data) VALUES `
				dataValues := []interface{}{}

				for j := 0; j < *dataFlag; j++ {
					dataQuery += fmt.Sprintf("($%d, $%d),", j*2+1, j*2+2)
					dataValues = append(dataValues, userId, Data)
				}

				dataQuery = dataQuery[:len(dataQuery)-1] // Remove the trailing comma

				_, err = db.Exec(dataQuery, dataValues...)
				if err != nil {
					log.Fatal(err)
				}
				bar.Add(1)
			}
		}()
	}

	for i := 0; i < *usersFlag; i++ {
		tasks <- i
	}
	close(tasks)

	wg.Wait()

	fmt.Println("All tasks have been completed successfully.")
}
