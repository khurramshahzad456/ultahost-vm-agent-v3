// internal/websocket/pending.go
package websocket

import (
	"fmt"
	"sync"
	"time"
)

// pendingEntry holds a channel and the agent identity that owns the task
type pendingEntry struct {
	ch            chan TaskResult
	agentIdentity string
	created       time.Time
}

var (
	pendingMtx sync.Mutex
	pendingMap = map[string]*pendingEntry{} // taskID -> entry
)

// registerPending registers a pending channel for taskID and returns the channel
func registerPending(taskID, agentIdentity string) chan TaskResult {
	pendingMtx.Lock()
	defer pendingMtx.Unlock()
	ch := make(chan TaskResult, 1)
	pendingMap[taskID] = &pendingEntry{
		ch:            ch,
		agentIdentity: agentIdentity,
		created:       time.Now(),
	}
	return ch
}

// unregisterPending removes the pending entry (used on timeout or send failure)
func unregisterPending(taskID string) {
	pendingMtx.Lock()
	defer pendingMtx.Unlock()
	if e, ok := pendingMap[taskID]; ok {
		// best-effort drain/close
		select {
		case <-e.ch:
		default:
		}
		close(e.ch)
		delete(pendingMap, taskID)
	}
}

// resolvePending sends result to the waiting channel and returns true if someone was waiting.
func resolvePending(taskID string, result TaskResult) bool {
	pendingMtx.Lock()
	entry, ok := pendingMap[taskID]
	if ok {
		delete(pendingMap, taskID)
	}
	pendingMtx.Unlock()

	if !ok {
		return false
	}
	// best-effort non-blocking send
	select {
	case entry.ch <- result:
	default:
	}
	close(entry.ch)
	return true
}

// waitForPending waits on the channel for taskID with timeout.
// It returns an error if no pending was registered or on timeout.
func waitForPending(taskID string, timeout time.Duration) (TaskResult, error) {
	pendingMtx.Lock()
	entry, ok := pendingMap[taskID]
	pendingMtx.Unlock()
	if !ok {
		return TaskResult{}, fmt.Errorf("no pending task: %s", taskID)
	}

	select {
	case res := <-entry.ch:
		return res, nil
	case <-time.After(timeout):
		// timeout, cleanup
		unregisterPending(taskID)
		return TaskResult{}, fmt.Errorf("timeout waiting for task result: %s", taskID)
	}
}

// failPendingForAgent fails all pending tasks that were intended for agentIdentity
// This is invoked on agent disconnect so callers don't wait forever.
func failPendingForAgent(agentIdentity, reason string) {
	pendingMtx.Lock()
	toFail := make(map[string]*pendingEntry)
	for id, e := range pendingMap {
		if e.agentIdentity == agentIdentity {
			toFail[id] = e
			delete(pendingMap, id)
		}
	}
	pendingMtx.Unlock()

	for id, e := range toFail {
		res := TaskResult{
			TaskID:       id,
			Task:         "",
			ExitCode:     -1,
			Stdout:       "",
			Stderr:       "agent disconnected or connection lost: " + reason,
			StartedAt:    time.Now().UTC().Format(time.RFC3339Nano),
			FinishedAt:   time.Now().UTC().Format(time.RFC3339Nano),
			DurationSec:  0,
			ChrootUsed:   false,
			CgroupUsed:   false,
			SignatureOK:  false,
			ScriptSHA256: "",
		}
		select {
		case e.ch <- res:
		default:
		}
		close(e.ch)
	}
}
