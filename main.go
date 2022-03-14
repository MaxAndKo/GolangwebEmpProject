package main

import (
	"empProject/controller"
	"log"
	"net/http"
	"os"
)

func setHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", controller.Home)
	mux.HandleFunc("/employee", controller.ShowEmp)
	mux.HandleFunc("/employee/list", controller.ShowEmps)
	mux.HandleFunc("/employee/create", controller.CreateEmp)
	mux.HandleFunc("/employee/remove", controller.RemoveEmp)
	mux.HandleFunc("/employee/update", controller.UpdateEmp)
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
