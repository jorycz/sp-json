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
	jsonValue := p.GetValueOfJsonKeyOptionallyUnderAnotherKey(payload, "properties", "allOf")
	if jsonValue != nil {
		slog.Info("found JSON - KEY", "data", jsonValue)
	} else {
		slog.Error("NOT found JSON - KEY", "data", jsonValue)
	}

	// Find key value using path
	jsonValueFromPath := p.GetValueOfJsonKeyOnPath(payload, []string{"properties", "nc:Vehicle", "oneOf", "[0]"})
	if jsonValueFromPath != nil {
		slog.Info("found JSON - PATH", "data", jsonValueFromPath)
	} else {
		slog.Error("NOT found JSON - PATH", "data", jsonValueFromPath)
	}

	// Get raw JSON
  	// rawJson := p.GetRawJson(payload)
  	// slog.Info("JSON - RAW", "data", rawJson)
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
