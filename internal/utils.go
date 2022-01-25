package internal

import (
	"encoding/json"
	"os"
)

// ParseFromJSONFile parses a JSON file to a struct.
func ParseFromJSONFile(path string, v interface{}) error {
	data, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	return json.NewDecoder(data).Decode(v)
}
