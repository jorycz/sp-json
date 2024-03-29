package parser

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
)

var (
	underKeyFound map[string]bool
	allKeyValues  []interface{}

	directPath []string
)

type JSONParser struct {
	lock sync.Mutex
}


func (p *JSONParser) GetValueOfJsonKeyOnPath(jsonData []byte, jsonPath []string) (interface{}, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	directPath = jsonPath
	dataMap, err := createMapFromJsonData(jsonData)
	if dataMap != nil {
		if val, ok := searchKeyInPath(dataMap); ok {
			return val, err
		}
	}
	slog.Debug("JSON value not found", "path", strings.Join(jsonPath, " > "))
	return nil, err
}

func (p *JSONParser) GetRawJson(jsonData []byte) (interface{}, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	result, err := unmarshalData(jsonData)
	return *result, err
}

// somewhereUnderThisAnotherKey - is optional. Use "" (empty string) if you want to search whole JSON for key.
// Always returns an array if any result is found
func (p *JSONParser) GetValueOfJsonKeyOptionallyUnderAnotherKey(jsonData []byte, jsonKey string, somewhereUnderThisAnotherKey string) (interface{}, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	dataMap, err := createMapFromJsonData(jsonData)
	if dataMap != nil {
		underKeyFound = make(map[string]bool)
		allKeyValues = make([]interface{}, 0)
		searchKeyAnywhereInMap(jsonKey, somewhereUnderThisAnotherKey, dataMap)
		if len(allKeyValues) > 0 {
			// Always return array
			return allKeyValues, err
		}
	}
	slog.Debug("JSON value not found", "key", jsonKey)
	return nil, err
}


// HELPERS //
// HELPERS //
// HELPERS //


// Search for value on particular JSON path
// Helper for GetValueOfJsonKeyOnPath
func searchKeyInPath(jsonData interface{}) (interface{}, bool) {
	// Final interface to return, if any
	var finalKeyValue interface{}

	// Loop through string array like []string{"Person", "[2]", "Count"} to get JSON value in that path
	// Get JSON PATH parts:
	for _, keyToSearch := range directPath {
		// Remove result found in previous loop if any
		finalKeyValue = nil

		// Check if current value in jsonData is a map
		if isMap, ok := jsonData.(map[string]interface{}); ok {
			// IF yes, check there is key I'm looking for with some value
			value, valueExists := isMap[keyToSearch]
			if valueExists {
				// In case it was last key I was looking for
				finalKeyValue = value
				// In case I need to process deeper to structure
				jsonData = value
			}
		}

		// Check if current value in JSON PATH should be an array index and remove [] around number
		replacer := strings.NewReplacer("[", "", "]", "")
		idx := replacer.Replace(keyToSearch)
		// Check if current keyToSearch is really number
		number, err := strconv.Atoi(idx)
		if err == nil {
			// Check if current jsonData is an array ...
			if arr, ok := jsonData.([]interface{}); ok {
				// ... and is full of joy
				if len(arr) > number {
					finalKeyValue = arr[number]
					jsonData = arr[number]
				}
			}
		}
	}

	if finalKeyValue != nil {
		return finalKeyValue, true
	}
	return nil, false
}


// Recursive function to search for a key in a nested map. It is possible specify required root key which could be somewhere in the middle
// Helper for GetValueOfJsonKeyOptionallyUnderAnotherKey
func searchKeyAnywhereInMap(key string, mustBeUnderKey string, dataMap map[string]interface{}) {
	// Check if the key exists in the current level of the map
	if value, ok := dataMap[key]; ok {
		if len(mustBeUnderKey) > 0 {
			if _, ok := underKeyFound[mustBeUnderKey]; ok {
				allKeyValues = append(allKeyValues, value)
			}
		} else {
			allKeyValues = append(allKeyValues, value)
		}
	}
	// Iterate over the values in the map
	for k, value := range dataMap {
		// Set current root key - reason I'm using map is that I want to get all keys I'm searching for from this level
		underKeyFound[k] = true
		// Check if the value is a nested map
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recursively search the nested map for the key
			searchKeyAnywhereInMap(key, mustBeUnderKey, nestedMap)
		}
		if nestedArray, ok := value.([]interface{}); ok {
			// Recursively search the nested array for the key
			searchKeyAnywhereInArray(key, mustBeUnderKey, nestedArray)
		}
		delete(underKeyFound, k)
	}
}

// Helper for searchKeyAnywhereInMap
func searchKeyAnywhereInArray(key string, mustBeUnderKey string, dataArray []interface{}) {
	for _, value := range dataArray {
		if nestedArray, ok := value.([]interface{}); ok {
			searchKeyAnywhereInArray(key, mustBeUnderKey, nestedArray)
		}
		if nestedMap, ok := value.(map[string]interface{}); ok {
			searchKeyAnywhereInMap(key, mustBeUnderKey, nestedMap)
		}
	}
}

// General helpers

// Prepare JSON data always as a map
// If it's not a map insert content of JSON to key k = { value }
func createMapFromJsonData(data []byte) (map[string]interface{}, error) {
	// JSON can hold:
	// {}
	// []
	// number, string
	// false, null, true

	result, err := unmarshalData(data)
	// I have a hack here - some JSON starts with [] (array) or special strings from JSON standard, so create map here to start processing always as a map
	if result != nil {
		tmpMap := make(map[string]interface{})
		switch v := (*result).(type) {
		case map[string]interface{}:
			slog.Debug("DATA", "type map", v)
			if dataMap, ok := (*result).(map[string]interface{}); ok {
				return dataMap, nil
			}
		case []interface{}:
			slog.Debug("JSON data", "type array", v)
			if dataMap, ok := (*result).([]interface{}); ok {
				tmpMap["k"] = dataMap
				// Insert fake key to requested jsonPath
				directPath = append([]string{"k"}, directPath...)
				return tmpMap, nil
			}
		case float64:
			slog.Debug("JSON data", "type number", v)
			if dataMap, ok := (*result).(float64); ok {
				tmpMap["k"] = dataMap
				return tmpMap, nil
			}
		case string:
			slog.Debug("JSON data", "type string", v)
			if dataMap, ok := (*result).(string); ok {
				tmpMap["k"] = dataMap
				return tmpMap, nil
			}
		case bool:
			slog.Debug("JSON data", "type bool", v)
			if dataMap, ok := (*result).(bool); ok {
				tmpMap["k"] = dataMap
				return tmpMap, nil
			}
		case nil:
			slog.Error("JSON data error", "type nil", v)
			tmpMap["k"] = "null"
			return tmpMap, nil
		default:
			err = fmt.Errorf("JSON data type not recognized. Type: %v", fmt.Sprintf("%T", *result))
		}
	}
	return nil, err
}

// Get raw JSON string
func unmarshalData(data []byte) (*interface{}, error) {
	var result interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		slog.Error("unmarshal JSON data error", "data", string(data[:]))
		return nil, err
	}
	return &result, nil
}
