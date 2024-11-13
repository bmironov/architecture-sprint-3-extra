package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	kafka "github.com/segmentio/kafka-go"
)

type LIGHTtelemetry struct {
	Id            int64     `json:"id"`
	LightId       int       `json:"light_id"`
	CreatedAt     time.Time `json:"created_at"`
	CurrentBright float32   `json:"current_bright"`
	TargetBright  float32   `json:"target_bright"`
}

var (
	db     *sql.DB
	dbUser string = "hvac_user"
	dbPass string = "hvac_password"
	dbHost string = "127.0.0.1"
	dbPort string = "5432"
	dbName string = "warm_home_light"

	kafkaURL    string = "localhost:9092"
	kafkaTopic  string = "data_topic"
	kafkaReader *kafka.Reader

	healthCheckPort string = "9090"
)

func envVar(value string, variable string) string {
	tmp, flag := os.LookupEnv(variable)
	if flag {
		//fmt.Printf("ENV %s = %s\n", variable, tmp)
		return tmp
	} else {
		//fmt.Printf("ENV %s = %s\n", variable, value)
		return value
	}
}

func initEnv() {
	dbUser = envVar(dbUser, "DB_USER")
	dbPass = envVar(dbPass, "DB_PASS")
	dbHost = envVar(dbHost, "DB_HOST")
	dbPort = envVar(dbPort, "DB_PORT")
	dbName = envVar(dbName, "DB_NAME")

	kafkaURL = envVar(kafkaURL, "KAFKA_URL")
	kafkaTopic = envVar(kafkaTopic, "KAFKA_TOPIC")

	healthCheckPort = envVar(healthCheckPort, "HEALTHCHECK_PORT")
}

func saveLightTelemetry(data *LIGHTtelemetry) error {
	sql := `INSERT INTO lights_telemetry(light_id, created_at, current_brightness, target_brightness)
	VALUES($1, $2, $3, $4) RETURNING light_telemetry_id`

	rows, err := db.Query(sql, data.LightId, data.CreatedAt, data.CurrentBright, data.TargetBright)
	if err != nil {
		return fmt.Errorf("can't execute INSERT INTO lights_telemetry: %w", err)
	}

	for rows.Next() {
		err = rows.Scan(&data.Id)
		if err != nil {
			return fmt.Errorf("can't find new light_telemetry_id: %w", err)
		}
	}

	return nil
}

func main() {
	var err error
	var data LIGHTtelemetry
	initEnv()

	// Connect to DB
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Start health-check listener
	go func() {
		fmt.Printf("Starting healthcheck endpoint at port %s\n", healthCheckPort)
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})
		_ = http.ListenAndServe(":"+healthCheckPort, nil)
	}()

	// Start Kafka Consumer
	logger := log.New(os.Stdout, "Kafka consumer: ", 0)
	kafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(kafkaURL, ","),
		Topic:    kafkaTopic,
		Logger:   logger,
		MinBytes: 10240,    // 10KB
		MaxBytes: 10485760, // 10MB
	})
	defer kafkaReader.Close()

	fmt.Printf("Starting to process messages from Kafka's %s topic '%s'\n",
		kafkaURL, kafkaTopic)
	for {
		msg, err := kafkaReader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalln(err)
		}

		err = json.Unmarshal(msg.Value, &data)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("Processing message light_id=%d, current: %5.2f target: %5.2f ... ",
			data.LightId, data.CurrentBright, data.TargetBright)
		err = saveLightTelemetry(&data)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("new ID=%d\n", data.Id)
	}
}
