package hvac

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	NOT_FOUND_ID int = -1
)

type HVAC struct {
	Id     int    `json:"id"`
	Model  string `json:"model"`
	Serial int64  `json:"serial_id"`
}

type HVACtelemetry struct {
	Id          int64     `json:"id"`
	HvacId      int       `json:"hvac_id"`
	CreatedAt   time.Time `json:"created_at"`
	CurrentTemp float32   `json:"current_temp"`
	TargetTemp  float32   `json:"target_temp"`
}

type HVACstate struct {
	State      string  `json:"state"`
	TargetTemp float32 `json:"target_temp"`
}

var (
	ALLOWED_STATES = []string{"on", "off"}
)

func CreateHvac(db *sql.DB, data *HVAC) (*HVAC, error) {
	sql := "INSERT INTO hvacs(model, serial_) VALUES($1, $2) RETURNING hvac_id"

	rows, err := db.Query(sql, data.Model, data.Serial)
	if err != nil {
		return nil, fmt.Errorf("can't execute INSERT INTO hvacs: %w", err)
	}

	for rows.Next() {
		err = rows.Scan(&data.Id)
		if err != nil {
			return nil, fmt.Errorf("can't find new hvac_id: %w", err)
		}
	}

	return data, nil
}

func UpdateHvac(db *sql.DB, data *HVAC) (*HVAC, error) {
	sql := "UPDATE hvacs SET model = $1, serial_ = $2 WHERE hvac_id = $3"

	rows, err := db.Exec(sql, data.Model, data.Serial, data.Id)
	if err != nil {
		return nil, fmt.Errorf("can't execute UPDATE hvacs: %w", err)
	}

	count, err := rows.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("error during UPDATE.RowsAffected hvacs: %w", err)
	}
	if count == 0 {
		data.Id = NOT_FOUND_ID
		return data, fmt.Errorf("no row hvacs.hvac_id = %d", data.Id)
	}

	return data, nil
}

func DeleteHvac(db *sql.DB, id int) (int, error) {
	sql := "DELETE FROM hvacs WHERE hvac_id = $1"

	rows, err := db.Exec(sql, id)
	if err != nil {
		return 0, fmt.Errorf("can't retrieve data about HVAC: %w", err)
	}

	count, err := rows.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error retrieving RowsAffected: %w", err)
	}
	if count == 0 {
		return -1 * id, fmt.Errorf("HVAC with hvac_id=%d not found", id)
	}

	return id, nil
}

func FindHvac(db *sql.DB, id int) (*HVAC, error) {
	var data HVAC
	sql := "SELECT hvac_id, model, serial_ FROM hvacs WHERE hvac_id = $1"

	rows, err := db.Query(sql, id)
	if err != nil {
		return nil, fmt.Errorf("can't retrieve data about HVAC: %w", err)
	}

	n := 0
	for rows.Next() {
		n++
		err = rows.Scan(&data.Id, &data.Model, &data.Serial)
		if err != nil {
			return nil, fmt.Errorf("can't process row from hvacs table: %w", err)
		}
	}

	if n == 0 {
		data.Id = NOT_FOUND_ID
		return &data, fmt.Errorf("can't find record with hvac_id = %d", id)
	}

	return &data, nil
}

func CreateHvacTelemetry(db *sql.DB, data *HVACtelemetry) (*HVACtelemetry, error) {
	sql := `INSERT INTO hvacs_telemetry(hvac_id, created_at, current_temperature, target_temperature)
            VALUES($1, $2, $3, $4) RETURNING hvac_telemetry_id`

	rows, err := db.Query(sql, data.HvacId, data.CreatedAt, data.CurrentTemp, data.TargetTemp)
	if err != nil {
		return nil, fmt.Errorf("can't execute INSERT INTO hvacs_telemetry: %w", err)
	}

	for rows.Next() {
		err = rows.Scan(&data.Id)
		if err != nil {
			return nil, fmt.Errorf("can't find new hvac_telemetry_id: %w", err)
		}
	}

	return data, nil
}

func GetHvacTelemetry(db *sql.DB, hvacId int) (data []HVACtelemetry, err error) {
	sql := `SELECT hvac_telemetry_id, hvac_id, created_at, current_temperature, target_temperature
            FROM hvacs_telemetry
            WHERE hvac_id = $1
            ORDER BY created_at DESC
            LIMIT 100`

	rows, err := db.Query(sql, hvacId)
	if err != nil {
		return nil, fmt.Errorf("can't execute SELECT FROM hvacs_telemetry: %w", err)
	}

	data = make([]HVACtelemetry, 0)
	for rows.Next() {
		var row HVACtelemetry
		err = rows.Scan(&row.Id, &row.HvacId, &row.CreatedAt, &row.CurrentTemp, &row.TargetTemp)
		if err != nil {
			return nil, fmt.Errorf("can't find new hvac_telemetry_id: %w", err)
		}
		data = append(data, row)
	}

	return data, nil
}
