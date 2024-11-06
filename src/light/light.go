package light

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	NOT_FOUND_ID int = -1
)

type LIGHT struct {
	Id     int    `json:"id"`
	Model  string `json:"model"`
	Serial int64  `json:"serial_id"`
}

type LIGHTtelemetry struct {
	Id            int64     `json:"id"`
	LightId       int       `json:"light_id"`
	CreatedAt     time.Time `json:"created_at"`
	CurrentBright float32   `json:"current_bright"`
	TargetBright  float32   `json:"target_bright"`
}

type LIGHTstate struct {
	State        string  `json:"state"`
	TargetBright float32 `json:"target_bright"`
}

var (
	ALLOWED_STATES = []string{"on", "off"}
)

func CreateLight(db *sql.DB, data *LIGHT) (*LIGHT, error) {
	sql := "INSERT INTO lights(model, serial_) VALUES($1, $2) RETURNING light_id"

	rows, err := db.Query(sql, data.Model, data.Serial)
	if err != nil {
		return nil, fmt.Errorf("can't execute INSERT INTO lights: %w", err)
	}

	for rows.Next() {
		err = rows.Scan(&data.Id)
		if err != nil {
			return nil, fmt.Errorf("can't find new light_id: %w", err)
		}
	}

	return data, nil
}

func UpdateLight(db *sql.DB, data *LIGHT) (*LIGHT, error) {
	sql := "UPDATE lights SET model = $1, serial_ = $2 WHERE light_id = $3"

	rows, err := db.Exec(sql, data.Model, data.Serial, data.Id)
	if err != nil {
		return nil, fmt.Errorf("can't execute UPDATE lights: %w", err)
	}

	count, err := rows.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("error during UPDATE.RowsAffected lights: %w", err)
	}
	if count == 0 {
		data.Id = NOT_FOUND_ID
		return data, fmt.Errorf("no row lights.light_id = %d", data.Id)
	}

	return data, nil
}

func DeleteLight(db *sql.DB, id int) (int, error) {
	sql := "DELETE FROM lights WHERE light_id = $1"

	rows, err := db.Exec(sql, id)
	if err != nil {
		return 0, fmt.Errorf("can't retrieve data about LIGHT: %w", err)
	}

	count, err := rows.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error retrieving RowsAffected: %w", err)
	}
	if count == 0 {
		return -1 * id, fmt.Errorf("LIGHT with light_id=%d not found", id)
	}

	return id, nil
}

func FindLight(db *sql.DB, id int) (*LIGHT, error) {
	var data LIGHT
	sql := "SELECT light_id, model, serial_ FROM lights WHERE light_id = $1"

	rows, err := db.Query(sql, id)
	if err != nil {
		return nil, fmt.Errorf("can't retrieve data about LIGHT: %w", err)
	}

	n := 0
	for rows.Next() {
		n++
		err = rows.Scan(&data.Id, &data.Model, &data.Serial)
		if err != nil {
			return nil, fmt.Errorf("can't process row from lights table: %w", err)
		}
	}

	if n == 0 {
		data.Id = NOT_FOUND_ID
		return &data, fmt.Errorf("can't find record with light_id = %d", id)
	}

	return &data, nil
}

func CreateLightTelemetry(db *sql.DB, data *LIGHTtelemetry) (*LIGHTtelemetry, error) {
	sql := `INSERT INTO lights_telemetry(light_id, created_at, current_brightness, target_brightness)
            VALUES($1, $2, $3, $4) RETURNING light_telemetry_id`

	rows, err := db.Query(sql, data.LightId, data.CreatedAt, data.CurrentBright, data.TargetBright)
	if err != nil {
		return nil, fmt.Errorf("can't execute INSERT INTO lights_telemetry: %w", err)
	}

	for rows.Next() {
		err = rows.Scan(&data.Id)
		if err != nil {
			return nil, fmt.Errorf("can't find new light_telemetry_id: %w", err)
		}
	}

	return data, nil
}

func GetLightTelemetry(db *sql.DB, lightId int) (data []LIGHTtelemetry, err error) {
	sql := `SELECT light_telemetry_id, light_id, created_at, current_brightness, target_brightness
            FROM lights_telemetry
            WHERE light_id = $1
            ORDER BY created_at DESC
            LIMIT 100`

	rows, err := db.Query(sql, lightId)
	if err != nil {
		return nil, fmt.Errorf("can't execute SELECT FROM lights_telemetry: %w", err)
	}

	data = make([]LIGHTtelemetry, 0)
	for rows.Next() {
		var row LIGHTtelemetry
		err = rows.Scan(&row.Id, &row.LightId, &row.CreatedAt, &row.CurrentBright, &row.TargetBright)
		if err != nil {
			return nil, fmt.Errorf("can't find new light_telemetry_id: %w", err)
		}
		data = append(data, row)
	}

	return data, nil
}
