package main

import (
	"database/sql"
	"fmt"
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
	mux.HandleFunc("/tables/create", checkTables)
}

func connectDB() (db *sql.DB, err error) {
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	return
}

func checkTables(w http.ResponseWriter, r *http.Request) {
	db, err := connectDB()
	if err != nil {
		log.Fatal("Ошибка открытия соединения с БД")
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS public.departments( name character varying(20) COLLATE pg_catalog.\"default\" NOT NULL, phone character varying(12) COLLATE pg_catalog.\"default\", CONSTRAINT departments_pkey PRIMARY KEY (name))")
	if err != nil {
		log.Println("Ошибка создания таблицы департаментов")
		http.Error(w, "Внутрення ошибка сервера", 500)
	}
	fmt.Fprintf(w, "Таблица создана")
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
