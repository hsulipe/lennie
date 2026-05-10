package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/hsulipe/lennie/login/internal/adapters/httphandler"
	pgadapter "github.com/hsulipe/lennie/login/internal/adapters/postgres"
	"github.com/hsulipe/lennie/login/internal/services"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func connectDatabase() *gorm.DB {
	dsn := "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable"
	db, err := gorm.Open(
		gormpostgres.Open(fmt.Sprintf(
			dsn,
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)),
		&gorm.Config{},
	)
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	db := connectDatabase()

	repo := pgadapter.NewUserRepository(db)
	signInSvc := services.NewSignInService(repo)
	signUpSvc := services.NewSignUpService(repo)

	handler := httphandler.NewHandler(signInSvc, signUpSvc, db != nil)
	router := httphandler.NewRouter(handler)

	fmt.Println("Server is running on http://localhost:" + port)
	http.ListenAndServe(":"+port, router)
}
