package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bootkemp-dev/datacat-backend/config"
	_ "github.com/lib/pq"
)

var db *sql.DB

func Connect() {

	c, err := config.NewConfig("./config.yml")
	if err != nil {
		log.Fatal(err)
	}

	psqlInfo := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name)
	log.Println(psqlInfo)

	database, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("sql.Open failed: %v\n", err)
	}

	err = database.Ping()
	if err != nil {
		log.Fatalf("database.Ping failed: %v\n", err)
	}

	db = database
	log.Println("[+] Connected to the database")
}
