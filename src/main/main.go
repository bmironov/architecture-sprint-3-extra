package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"database/sql"

	_ "github.com/lib/pq"
)

var (
	db      *sql.DB
	appPort string = "8080"
	dbUser  string = "hvac_user"
	dbPass  string = "hvac_password"
	dbHost  string = "127.0.0.1"
	dbPort  string = "5432"
	dbName  string = "warm_home_hvac"
	dbFlag  bool   = false
)

func envVar(value string, variable string) string {
	tmp, flag := os.LookupEnv(variable)
	if flag {
		return tmp
	} else {
		return value
	}
}

func initEnv() {
	appPort = envVar(appPort, "APP_PORT")
	dbUser = envVar(appPort, "DB_USER")
	dbPass = envVar(appPort, "DB_PASS")
	dbHost = envVar(appPort, "DB_HOST")
	dbPort = envVar(appPort, "DB_PORT")
	dbName = envVar(appPort, "DB_NAME")
}

func main() {
	var err error

	initEnv()
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

	mux.HandleFunc("POST   /lights", createLight)
	mux.HandleFunc("PUT    /lights/{id}", updateLight)
	mux.HandleFunc("GET    /lights/{id}", findLight)
	mux.HandleFunc("DELETE /lights/{id}", deleteLight)
	mux.HandleFunc("GET    /lights/telemetry/{id}", getLightTelemetry)
	mux.HandleFunc("PUT    /lights/telemetry/{id}", addLightTelemetry)
	mux.HandleFunc("PUT    /lights/state/{id}", sendLightCommand)

	mux.HandleFunc("GET /", hello)

	log.Fatal(http.ListenAndServe("127.0.0.1:"+appPort, mux))
}
