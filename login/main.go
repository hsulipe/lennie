package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, `{"healthy": true}`)
}

func main() {
	r := mux.NewRouter()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	r.HandleFunc("/healthz", HealthCheckHandler).Methods(http.MethodGet)

	fmt.Println("Server is running on http://localhost:" + port)
	http.ListenAndServe(":"+port, r)
}
