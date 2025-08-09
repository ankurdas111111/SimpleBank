package main

import (
	"database/sql"
	"log"

	"github.com/ankurdas111111/simplebank/api"
	db "github.com/ankurdas111111/simplebank/db/sqlc"
	"github.com/ankurdas111111/simplebank/util"
	_ "github.com/lib/pq"
)


func main(){

	config,err := util.LoadConfig(".") // . means current folder
	if err != nil{
		log.Fatal("Can not load config:",err)
	}

	conn, err := sql.Open(config.DBdriver, config.DBsource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil{
		log.Fatal("Can not create server:", err)
	}
	err = server.Start(config.ServerAddress)
	if err != nil{
		log.Fatal("Can not start the server:", err)
	}
}