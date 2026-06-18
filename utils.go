package main

import "encoding/json"

// Packet is the object received on the TCP connection
type Packet struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

// InitContent starts a new recording session for a given cat (device id)
type InitContent struct {
	Cat       string `json:"cat"`
	Timestamp string `json:"timestamp"`
	Context   string `json:"context"`
}

// LogContent carries a single text log line in ESP-IDF format
type LogContent struct {
	Data string `json:"data"`
}

// RawContent carries a single chunk of raw sensor data.
type RawContent struct {
	TimeMs int    `json:"time_ms"`
	Data   string `json:"data"`
}
