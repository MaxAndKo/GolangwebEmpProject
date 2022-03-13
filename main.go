package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

func setHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", home)
	mux.HandleFunc("/employee", showEmp)
	mux.HandleFunc("/employee/list", showEmps)
	mux.HandleFunc("/employee/create", createEmp)
	mux.HandleFunc("/employee/remove", removeEmp)
	mux.HandleFunc("/employee/update", updateEmp)
}

func connectDB() (db *sql.DB, err error) {
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	return
}

func main() {
	mux := http.NewServeMux()
	setHandlers(mux)

	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	log.Println("Start server")
	err := http.ListenAndServe(":"+port, mux)
	log.Fatal(err)
}
