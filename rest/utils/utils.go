package utils

import (
	"bytes"
	"encoding/json"
	"os"
)

func WriteJsonFile(filepath string, data []byte) (err error) {
	var out bytes.Buffer
	json.Indent(&out, data, "", "\t")
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}
	out.WriteTo(f)
	defer f.Close()
	return
}
