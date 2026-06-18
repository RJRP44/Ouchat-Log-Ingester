package main

import (
	"database/sql"
	"time"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(db *sql.DB) *Database {
	return &Database{db: db}
}

// CreateCat inserts the cat (device id) if it doesn't already exist.
func (r *Database) CreateCat(cat string) error {
	_, err := r.db.Exec(
		`INSERT INTO cats (cat) VALUES ($1) ON CONFLICT (cat) DO NOTHING`,
		cat,
	)
	return err
}

// CreateSession inserts a new session (cat, timestamp) if it doesn't already exist.
func (r *Database) CreateSession(cat string, ts time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO sessions (cat, timestamp) VALUES ($1, $2) ON CONFLICT (cat, timestamp) DO NOTHING`,
		cat, ts,
	)
	return err
}

// InsertCalibration stores the calibration payload sent in an init message.
// If the same (cat, timestamp) session sends an init message again, the
// calibration payload is overwritten rather than rejected.
func (r *Database) InsertCalibration(cat string, ts time.Time, data string) error {
	_, err := r.db.Exec(
		`INSERT INTO calibration_data (cat, timestamp, data) VALUES ($1, $2, $3)`,
		cat, ts, data,
	)
	return err
}

// InsertLog stores one log line for an active session. mcuMs is nil when it
// couldn't be parsed from the log line, since the column is nullable and
// the CHECK (mcu_ms > 0) constraint only applies to non-null values.
func (r *Database) InsertLog(cat string, ts time.Time, mcuMs *int, level int, message string) error {
	_, err := r.db.Exec(
		`INSERT INTO logs (cat, timestamp, mcu_ms, level, log) VALUES ($1, $2, $3, $4, $5)`,
		cat, ts, nullableInt(mcuMs), level, message,
	)
	return err
}

// InsertRaw stores one raw sensor data chunk for an active session.
func (r *Database) InsertRaw(cat string, ts time.Time, mcuMs int, data string) error {
	_, err := r.db.Exec(
		`INSERT INTO raw_data (cat, timestamp, mcu_ms, data) VALUES ($1, $2, $3, $4)`,
		cat, ts, mcuMs, data,
	)
	return err
}

func nullableInt(v *int) interface{} {
	if v == nil {
		return nil
	}
	return *v
}
