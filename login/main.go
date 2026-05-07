package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db_connected = false

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	io.WriteString(w, `{"healthy": true, "db_connected": `+fmt.Sprintf("%t", db_connected)+`}`)
}

func ConnectDatabase() *gorm.DB {
	dsn := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable"
	db, err := gorm.Open(
		postgres.Open(
			fmt.Sprintf(
				dsn,
				os.Getenv("DB_HOST"),
				os.Getenv("DB_USER"),
				os.Getenv("DB_PASSWORD"),
				os.Getenv("DB_NAME"),
				os.Getenv("DB_PORT"),
			),
		),
		&gorm.Config{},
	)
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func main() {
	r := mux.NewRouter()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	db := ConnectDatabase()
	if db != nil {
		db_connected = true
	}
	r.HandleFunc("/health", HealthCheckHandler).Methods(http.MethodGet)

	fmt.Println("Server is running on http://localhost:" + port)
	http.ListenAndServe(":"+port, r)
}
