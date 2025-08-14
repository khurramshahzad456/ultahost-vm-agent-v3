package runner

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

const defaultManifest = "scripts/manifest.json"

// Public entrypoint used by your WS/HTTP handler
func ExecuteSignedTaskJSON(reqJSON []byte) ([]byte, error) {
	os.MkdirAll("logs", os.ModePerm)
	logFile, err := os.OpenFile("logs/ultaai/agent.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	log.SetOutput(logFile)
	var req TaskRequest
	if err := json.Unmarshal(reqJSON, &req); err != nil {
		return nil, err
	}
	if req.Type != "task" {
		return nil, errors.New("unsupported message type")
	}
	res, err := ExecuteSignedTask(req, defaultManifest)
	if err != nil {
		return nil, err
	}
	return json.Marshal(res)
}
