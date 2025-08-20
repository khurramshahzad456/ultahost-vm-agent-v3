// internal/websocket/task.go
package websocket

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ultahost-ai-gateway/internal/utils"

	"github.com/google/uuid"
)

// TaskRequest is sent to the agent
type TaskRequest struct {
	Type      string   `json:"type"`    // "task"
	TaskID    string   `json:"task_id"` // server-generated UUID
	Task      string   `json:"task"`    // allowlist name e.g., "install_wordpress"
	Args      []string `json:"args,omitempty"`
	Timestamp string   `json:"timestamp"` // RFC3339
	Nonce     string   `json:"nonce"`
	Signature string   `json:"signature"` // base64 HMAC-SHA256
}

// TaskResult is the agent's response (expected JSON shape)
// Fields match the agent runner TaskResult shape.
type TaskResult struct {
	TaskID       string `json:"task_id"`
	Task         string `json:"task"`
	ExitCode     int    `json:"exit_code"`
	Stdout       string `json:"stdout"`
	Stderr       string `json:"stderr"`
	StartedAt    string `json:"started_at"`
	FinishedAt   string `json:"finished_at"`
	DurationSec  int64  `json:"duration_sec"`
	ChrootUsed   bool   `json:"chroot_used"`
	CgroupUsed   bool   `json:"cgroup_used"`
	SignatureOK  bool   `json:"signature_ok"`
	ScriptSHA256 string `json:"script_sha256"`
}

// canonicalString must exactly match the agent's canonical string for HMAC
func canonicalString(task string, args []string, nonce, ts string) string {
	return fmt.Sprintf("v1|%s|%s|%s|%s", task, strings.Join(args, " "), nonce, ts)
}

// SendSignedTask sends a signed task to the agent and returns the generated taskID.
// This does not wait for a result.
func SendSignedTask(vpsId string, task string, args []string) (string, error) {
	CN := "Agent_" + vpsId
	keyInfo, exist := utils.GetAgentKeys(CN)
	if !exist {
		return "", fmt.Errorf("no key info for %s", CN)
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	nonce := uuid.NewString()
	taskID := uuid.NewString()

	msg := canonicalString(task, args, nonce, ts)
	mac := hmac.New(sha256.New, []byte(keyInfo.SignatureSecret))
	mac.Write([]byte(msg))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	tr := TaskRequest{
		Type:      "task",
		TaskID:    taskID,
		Task:      task,
		Args:      args,
		Timestamp: ts,
		Nonce:     nonce,
		Signature: sig,
	}

	payload, err := json.Marshal(tr)
	if err != nil {
		return "", err
	}

	if err := SendMessage(vpsId, payload); err != nil {
		return "", err
	}
	return taskID, nil
}

// SendSignedTaskAndWait sends a signed task and waits up to `timeout` for a task_result from the agent.
// Returns the TaskResult or an error on send / timeout.
func SendSignedTaskAndWait(vpsId string, task string, args []string, timeout time.Duration) (TaskResult, error) {
	CN := "Agent_" + vpsId
	keyInfo, exist := utils.GetAgentKeys(CN)
	if !exist {
		return TaskResult{}, fmt.Errorf("no key info for %s", CN)
	}

	// ts := time.Now().UTC().Format(time.RFC3339)
	ts := time.Now().UTC().Format(time.RFC3339Nano)

	nonce := uuid.NewString()
	taskID := uuid.NewString()

	msg := canonicalString(task, args, nonce, ts)
	mac := hmac.New(sha256.New, []byte(keyInfo.SignatureSecret))
	mac.Write([]byte(msg))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	tr := TaskRequest{
		Type:      "task",
		TaskID:    taskID,
		Task:      task,
		Args:      args,
		Timestamp: ts,
		Nonce:     nonce,
		Signature: sig,
	}

	payload, err := json.Marshal(tr)
	if err != nil {
		return TaskResult{}, err
	}

	// register pending before send so we don't race with an immediate result
	ch := registerPending(taskID, keyInfo.IdentityToken)

	// try sending
	if err := SendMessage(vpsId, payload); err != nil {
		// cleanup pending and return
		unregisterPending(taskID)
		return TaskResult{}, fmt.Errorf("send message failed: %w", err)
	}

	// wait
	select {
	case res := <-ch:
		return res, nil
	case <-time.After(timeout):
		unregisterPending(taskID)
		return TaskResult{}, fmt.Errorf("timeout waiting for task result (task_id=%s)", taskID)
	}
}
