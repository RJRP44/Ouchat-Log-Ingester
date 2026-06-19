package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// timestampLayout matches the timestamp format used by postgres
const timestampLayout = "2006-01-02 15:04:05"
const electronicsTimestampLayout = "2006-01-02_15-04-05"

// connState holds the session context currently active on a connection.
// A connection is expected to send exactly one "init" message before any
// "log" or "raw" message, but re-init mid-connection is tolerated.
type connState struct {
	cat        string
	timestamp  time.Time
	hasSession bool
}

func closeConnection(conn net.Conn) {
	err := conn.Close()
	if err != nil {
		log.Printf("Error closing connection: %v", err)
	}
}

// handleConnection reads consecutive JSON messages from the connection and dispatches them
func handleConnection(conn net.Conn, db *Database) {

	defer closeConnection(conn)

	remote := conn.RemoteAddr()
	log.Printf("new connection from %s", remote)

	decoder := json.NewDecoder(conn)
	state := &connState{}

	for {
		var env Packet
		if err := decoder.Decode(&env); err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("connection closed by %s", remote)
			} else {
				log.Printf("decode error from %s: %v", remote, err)
			}
			return
		}
		if err := dispatch(&env, state, db); err != nil {
			log.Printf("error handling %q message from %s: %v", env.Type, remote, err)
		}
	}
}

func dispatch(env *Packet, state *connState, db *Database) error {
	switch env.Type {
	case "init":
		return handleInit(env, state, db)
	case "log":
		return handleLog(env, state, db)
	case "raw":
		return handleRaw(env, state, db)
	default:
		return fmt.Errorf("unknown message type %q", env.Type)
	}
}

func handleInit(env *Packet, state *connState, db *Database) error {
	var content InitContent
	if err := json.Unmarshal(env.Content, &content); err != nil {
		return fmt.Errorf("invalid init content: %w", err)
	}

	if content.Timestamp == "nocam" {
		loc, _ := time.LoadLocation("Europe/Paris")
		currentTime := time.Now().In(loc)
		content.Timestamp = currentTime.Format(timestampLayout)
	}

	ts, err := time.Parse(electronicsTimestampLayout, content.Timestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp %q: %w", content.Timestamp, err)
	}

	//Create the id only if this device is not registered
	if err := db.CreateCat(content.Cat); err != nil {
		return fmt.Errorf("ensure cat: %w", err)
	}
	if err := db.CreateSession(content.Cat, ts); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	if err := db.InsertCalibration(content.Cat, ts, content.Context); err != nil {
		return fmt.Errorf("insert calibration: %w", err)
	}

	state.cat = content.Cat
	state.timestamp = ts
	state.hasSession = true
	log.Printf("session started: cat=%s timestamp=%s", content.Cat, ts)
	return nil
}

func handleLog(env *Packet, state *connState, db *Database) error {
	if !state.hasSession {
		return errors.New("log message received before init")
	}

	var content LogContent
	if err := json.Unmarshal(env.Content, &content); err != nil {
		return fmt.Errorf("invalid log content: %w", err)
	}

	level, mcuMs, message := parseLogLine(content.Data)

	var mcuMsPtr *int
	if mcuMs > 0 {
		mcuMsPtr = &mcuMs
	}

	return db.InsertLog(state.cat, state.timestamp, mcuMsPtr, level, message)
}

func handleRaw(env *Packet, state *connState, db *Database) error {
	if !state.hasSession {
		return errors.New("raw sensor data received before init")
	}

	var content RawContent
	if err := json.Unmarshal(env.Content, &content); err != nil {
		return fmt.Errorf("invalid raw content: %w", err)
	}

	return db.InsertRaw(state.cat, state.timestamp, content.TimeMs, content.Data)
}
