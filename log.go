package main

import (
	"regexp"
	"strconv"
)

// espLogRegex matches ESP-IDF style log lines: I (56) boot: chip revision: v3.0
var espLogRegex = regexp.MustCompile(`(?m)^\s*([EWIDV]) \((\d+)\)\s*(.*)\r?$`)

// levelMap converts the ESP-IDF level letter into an integer level
var levelMap = map[byte]int{
	'E': 1, // Error
	'W': 2, // Warning
	'I': 3, // Info
	'D': 4, // Debug
	'V': 5, // Verbose
}

// parseLogLine extracts the level and the MCU uptime (in ms) and message
func parseLogLine(line string) (level int, mcuMs int, message string) {
	matches := espLogRegex.FindStringSubmatch(line)
	if matches == nil {
		//Return a default value for the line if it not correct
		return 0, 0, line
	}

	level = levelMap[matches[1][0]]

	parsedMs, err := strconv.Atoi(matches[2])
	if err == nil {
		mcuMs = parsedMs
	}

	message = matches[3]
	return level, mcuMs, message
}
