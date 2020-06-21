package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "vAn8ish!"
	dbName := "quizgame"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName+"?parseTime=true")
	if err != nil {
		panic(err.Error())
	}
	return db
}

func dbSelect() {

	db := dbConn()
	selDB, err := db.Query("SELECT id, name, date_created FROM games WHERE id=?", 1)
	if err != nil {
		panic(err.Error())
	}
	for selDB.Next() {
		var id int
		var name string
		var date_created *time.Time
		err = selDB.Scan(&id, &name, &date_created)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("Row %d name %s\n", id, name)
	}
	defer db.Close()
}
