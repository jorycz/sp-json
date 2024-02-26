package main

import (
	"os"
	"log/slog"

	"github.com/jorycz/sp-json"
)

func main() {
	
	p := &parser.JSONParser{}

	payload := testJsonData()

	// Find key value using full JSON search - optionally limit result with another key using underThisAnotherKey
	jsonValue, errKey := p.GetValueOfJsonKeyOptionallyUnderAnotherKey(payload, "properties", "allOf")
	if errKey != nil {
		slog.Error("ERROR", "method KEY", "msg", errKey)
	}
	if jsonValue != nil {
		slog.Info("found JSON - KEY", "data", jsonValue)
	} else {
		slog.Error("NOT found JSON - KEY", "data", jsonValue)
	}

	// Find key value using path
	jsonValueFromPath, errPath := p.GetValueOfJsonKeyOnPath(payload, []string{"properties", "nc:Vehicle", "oneOf", "[0]"})
	if errPath != nil {
		slog.Error("ERROR", "method PATH", "msg", errKey)
	}
	if jsonValueFromPath != nil {
		slog.Info("found JSON - PATH", "data", jsonValueFromPath)
	} else {
		slog.Error("NOT found JSON - PATH", "data", jsonValueFromPath)
	}

	// Get raw JSON
  	rawJson, errRaw := p.GetRawJson(payload)
	if errRaw != nil {
		slog.Error("ERROR", "method RAW", "msg", errKey)
	}
  	slog.Info("JSON - RAW", "data", rawJson)
}


func testJsonData() []byte {
	file := "test-files/complex.json"
	data, err := os.ReadFile(file)
	if err != nil {
		slog.Error("Error reading file:", err)
		return nil
	}
	return data
}
