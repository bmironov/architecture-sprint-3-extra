package main

import (
	"fmt"
	"log"
	"net/http"

	"database/sql"

	_ "github.com/lib/pq"
)

var (
	db      *sql.DB
	appPort string = "8080" // os.Getenv("APP_PORT")
	dbUser  string = "hvac_user"
	dbPass  string = "hvac_password"
	dbHost  string = "127.0.0.1"
	dbPort  string = "5432"
	dbName  string = "warm_home_hvac"
	dbFlag  bool   = false
)

func main() {
	var err error
	mux := http.NewServeMux()

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// k8s
	mux.HandleFunc("GET /livez", livez)
	mux.HandleFunc("GET /readyz", readyz)

	mux.HandleFunc("POST   /hvac", createHvac)
	mux.HandleFunc("PUT    /hvac/{id}", updateHvac)
	mux.HandleFunc("GET    /hvac/{id}", findHvac)
	mux.HandleFunc("DELETE /hvac/{id}", deleteHvac)
	mux.HandleFunc("GET    /hvac/telemetry/{id}", getHvacTelemetry)
	mux.HandleFunc("PUT    /hvac/telemetry/{id}", addHvacTelemetry)
	mux.HandleFunc("PUT    /hvac/state/{id}", sendHvacCommand)

	mux.HandleFunc("GET /", hello)

	log.Fatal(http.ListenAndServe("127.0.0.1:"+appPort, mux))
}
