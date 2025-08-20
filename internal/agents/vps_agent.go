// internal/agents/vps_agent.go
package agents

import (
	"fmt"
	"time"

	"ultahost-ai-gateway/internal/ai"
	"ultahost-ai-gateway/internal/pkg/models"
	"ultahost-ai-gateway/internal/websocket"
)

func HandleVPS(req *models.ChatRequest, functionList []string) (string, error) {
	functionName, err := ai.ClassifyFunctionWithinAgent(req.Message, functionList)
	if err != nil {
		return "", err
	}

	// require target VPSID for agent tasks
	vpsId := req.VPSID
	if vpsId == "" {
		return "", fmt.Errorf("vps_id is required to perform agent tasks; include it in your request")
	}

	// choose a reasonable backend wait timeout for result (adjustable)
	waitTimeout := 2 * time.Minute

	switch functionName {
	case "checkuptime":
		res, err := websocket.SendSignedTaskAndWait(vpsId, "check_uptime", req.Args, waitTimeout)
		if err != nil {
			return "", fmt.Errorf("dispatch/check_uptime failed: %w", err)
		}
		if res.ExitCode == 0 {
			return res.Stdout, nil
		}
		return fmt.Sprintf("Command failed (exit=%d): %s", res.ExitCode, res.Stderr), nil

	case "checkdiskspace":
		res, err := websocket.SendSignedTaskAndWait(vpsId, "check_diskspace", req.Args, waitTimeout)
		if err != nil {
			return "", fmt.Errorf("dispatch/check_diskspace failed: %w", err)
		}
		if res.ExitCode == 0 {
			return res.Stdout, nil
		}
		return fmt.Sprintf("Command failed (exit=%d): %s", res.ExitCode, res.Stderr), nil

	case "installwordpress", "install_wordpress":
		// install can take longer; choose a longer wait (adjust as needed)
		installWait := 10 * time.Minute
		res, err := websocket.SendSignedTaskAndWait(vpsId, "install_wordpress", req.Args, installWait)
		if err != nil {
			return "", fmt.Errorf("dispatch/install_wordpress failed: %w", err)
		}
		if res.ExitCode == 0 {
			return res.Stdout, nil
		}
		return fmt.Sprintf("Install failed (exit=%d): %s", res.ExitCode, res.Stderr), nil

	default:
		return "I couldn't match your request to a known VPS function.", nil
	}
}
