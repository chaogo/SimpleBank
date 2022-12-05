package main

import (
	"database/sql"
	"log"

	"github.com/chaogo/SimpleBank/api"
	db "github.com/chaogo/SimpleBank/db/sqlc"
	"github.com/chaogo/SimpleBank/util"
	_ "github.com/lib/pq"
)

// entry point for our server
func main() {
	// to create a server, we need to connect to the database and create a Store first
	config, err := util.LoadConfig(".") // path . represents the same location
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress) // listening
	if err != nil {
		log.Fatal("cannot start server:", err)
	}
}

