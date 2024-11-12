package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"database/sql"

	_ "github.com/lib/pq"
	kafka "github.com/segmentio/kafka-go"
)

var (
	db      *sql.DB
	appPort string = "8888"
	dbUser  string = "hvac_user"
	dbPass  string = "hvac_password"
	dbHost  string = "127.0.0.1"
	dbPort  string = "5432"
	dbName  string = "warm_home_hvac"
	dbFlag  bool   = false // true when DB responds to "Ping" call

	kafkaURL    string = "localhost:9092"
	kafkaTopic  string = "data_topic"
	kafkaWriter *kafka.Writer
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
	dbUser = envVar(dbUser, "DB_USER")
	dbPass = envVar(dbPass, "DB_PASS")
	dbHost = envVar(dbHost, "DB_HOST")
	dbPort = envVar(dbPort, "DB_PORT")
	dbName = envVar(dbName, "DB_NAME")

	kafkaURL = envVar(kafkaURL, "KAFKA_URL")
	kafkaTopic = envVar(kafkaTopic, "KAFKA_TOPIC")
}

func main() {
	var err error

	initEnv()
	mux := http.NewServeMux()

	// Connect to DB
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Init Kafka Producer
	logger := log.New(os.Stdout, "Kafka producer: ", 0)
	kafkaWriter = &kafka.Writer{
		Addr:        kafka.TCP(kafkaURL),
		Topic:       kafkaTopic,
		Balancer:    &kafka.LeastBytes{},
		Logger:      logger,
		Compression: kafka.Snappy,
	}
	defer kafkaWriter.Close()

	// k8s
	mux.HandleFunc("GET /livez", livez)
	mux.HandleFunc("GET /readyz", readyz)

	mux.HandleFunc("POST   /hvac", createHvac)
	mux.HandleFunc("PUT    /hvac/{id}", updateHvac)
	mux.HandleFunc("GET    /hvac/{id}", findHvac)
	mux.HandleFunc("DELETE /hvac/{id}", deleteHvac)
	mux.HandleFunc("GET    /hvac/{id}/telemetry", getHvacTelemetry)
	mux.HandleFunc("PUT    /hvac/{id}/telemetry", addHvacTelemetry)
	mux.HandleFunc("PUT    /hvac/{id}/state", sendHvacCommand)

	mux.HandleFunc("POST   /lights", createLight)
	mux.HandleFunc("PUT    /lights/{id}", updateLight)
	mux.HandleFunc("GET    /lights/{id}", findLight)
	mux.HandleFunc("DELETE /lights/{id}", deleteLight)
	mux.HandleFunc("GET    /lights/{id}/telemetry", getLightTelemetry)
	mux.HandleFunc("PUT    /lights/{id}/telemetry", addLightTelemetry)
	mux.HandleFunc("PUT    /lights/{id}/state", sendLightCommand)

	mux.HandleFunc("GET /", hello)

	log.Fatal(http.ListenAndServe(":"+appPort, mux))
}
