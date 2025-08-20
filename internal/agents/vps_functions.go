package agents

import (
	"fmt"
	"ultahost-ai-gateway/internal/ai"
	"ultahost-ai-gateway/internal/pkg/models"
)

// List of all available VPS functions
var VPSFunctionList = []string{
	"checkUptime",
	"checkDiskSpace",
	"installWordPress",
}

// Function to check system uptime
func checkUptime(req *models.ChatRequest) (string, error) {
	rawOutput := "Uptime: 5 days 3 hours"
	summary, err := ai.SummarizeResponse(rawOutput)

	if err != nil {
		// fallback to raw
		return rawOutput, nil
	}
	return summary, nil
}

// Function to check disk space usage
func checkDiskSpace(req *models.ChatRequest) (string, error) {
	rawOutput := "Filesystem /dev/sda1 has used 45% of total space. 55% is still available."
	summary, err := ai.SummarizeResponse(rawOutput)
	if err != nil {
		return rawOutput, nil
	}
	return summary, nil
}

// Function to simulate WordPress installation
func installWordPress(req *models.ChatRequest) (string, error) {
	rawOutput := `
Step 1: Updated package lists.
Step 2: Installed Apache, PHP.
Step 3: Installed and configured MySQL.
Step 4: Downloaded latest WordPress archive.
Step 5: Set permissions and restarted Apache.
Installation complete.
`
	summary, err := ai.SummarizeResponse(rawOutput)
	fmt.Println("************************************************", summary)

	if err != nil {
		return rawOutput, nil
	}
	return summary, nil
}
