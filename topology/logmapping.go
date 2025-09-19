package topology

import (
	"encoding/json"
	"io"
	"os"
)

type SimulatorLogInfo struct {
	Filename string `json:"filename"`
	FileType string `json:"type"`
}

type LogMapping struct {
	SimulatorLogs map[string]SimulatorLogInfo `json:"logs"`
}

func ParseLogMapping(filename string) (*LogMapping, error) {
	var lm LogMapping

	mapping_file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer mapping_file.Close()

	bytes, err := io.ReadAll(mapping_file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(bytes), &lm)
	if err != nil {
		return nil, err
	}

	return &lm, nil
}
