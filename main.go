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
	mux.HandleFunc("/employee/tables/create", checkTables)
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
		return
	}

	_, err = db.Exec("CREATE SEQUENCE IF NOT EXISTS public.employees_id_seq INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 OWNED BY employees.id")
	if err != nil {
		log.Println("Ошибка sequence для поля id таблицы работники")
		http.Error(w, "Внутрення ошибка сервера", 500)
		return
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS public.employees(id integer NOT NULL DEFAULT nextval('employees_id_seq'::regclass), name character varying(25) COLLATE pg_catalog.\"default\", surname character varying(25) COLLATE pg_catalog.\"default\", phone character varying(12) COLLATE pg_catalog.\"default\", company_id integer, department character varying(20) COLLATE pg_catalog.\"default\", CONSTRAINT employees_pkey PRIMARY KEY (id), CONSTRAINT department_fkey FOREIGN KEY (department) REFERENCES public.departments (name) MATCH SIMPLE ON UPDATE NO ACTION ON DELETE NO ACTION NOT VALID)")
	if err != nil {
		log.Println("Ошибка создания таблицы работников")
		http.Error(w, "Внутрення ошибка сервера", 500)
		return
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS public.passports (passport_number character varying(20) COLLATE pg_catalog.\"default\" NOT NULL, passport_type character varying(20) COLLATE pg_catalog.\"default\" NOT NULL, employee_id integer, CONSTRAINT passports_pkey PRIMARY KEY (passport_number), CONSTRAINT employee_fk FOREIGN KEY (employee_id) REFERENCES public.employees (id) MATCH SIMPLE ON UPDATE NO ACTION ON DELETE CASCADE NOT VALID)")
	if err != nil {
		log.Println("Ошибка создания таблицы паспортов")
		http.Error(w, "Внутрення ошибка сервера", 500)
		return
	}

	fmt.Fprintf(w, "Таблицы созданы")
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
