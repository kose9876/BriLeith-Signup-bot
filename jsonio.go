package main

import (
	"bytes"
	"encoding/json"
	"os"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

func readJSONFile(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	data = bytes.TrimPrefix(data, utf8BOM)
	return json.Unmarshal(data, value)
}
