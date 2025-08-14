package agent

import (
	"bytes"
	"context"
	"crypto/hmac"
	"math/rand"

	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
	"ultahost-agent/internal/runner"
	"ultahost-agent/utils"

	"github.com/gorilla/websocket"
)

var (
	configDir    = "./test-vm-agent"
	heartbeatCtr uint64 // monotonic counter
	heartbeatVer = "1"  // bump if you change signing format
	agentIDPath  = "/etc/ultaai-agent-id"
)

// ---- helpers ----

type heartbeatPayload struct {
	Type      string `json:"type"`      // "heartbeat"
	Version   string `json:"version"`   // schema/signature version
	AgentID   string `json:"agent_id"`  // unique agent id (UUID)
	Counter   uint64 `json:"counter"`   // anti-replay (monotonic per-process)
	Nonce     string `json:"nonce"`     // base64 random bytes
	Timestamp string `json:"timestamp"` // RFC3339
	Signature string `json:"signature"` // base64 HMAC-SHA256 over canonical string
}

// canonical string: fixed order & delimiter to avoid ambiguity
func canonicalString(v, agent string, ctr uint64, nonceB64, ts string) string {
	// Keep this EXACT order and delimiter consistent across client/server
	return fmt.Sprintf("%s|%s|%d|%s|%s", v, agent, ctr, nonceB64, ts)
}

func signHMACSHA256(secret []byte, msg string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(msg))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

const (
	// websocket timeouts
	pongWait   = 60 * time.Second
	writeWait  = 10 * time.Second
	pingPeriod = (pongWait * 9) / 10
	readLimit  = 1024 * 1024 // 1MB

	// reconnect/backoff
	initialBackoff = 1 * time.Second
	maxBackoff     = 30 * time.Second
	maxAttempts    = 0 // 0 means infinite attempts (until ctx canceled)
)

func ConnectAndMaintain(ctx context.Context) error {
	// seed rand once
	rand.Seed(time.Now().UnixNano())

	// Preload static assets used across attempts
	clientCert, err := tls.LoadX509KeyPair(configDir+"/client.crt", configDir+"/client.key")
	if err != nil {
		return fmt.Errorf("failed to load client cert/key: %w", err)
	}

	caCertPEM, err := os.ReadFile("./crts/ca.crt")
	if err != nil {
		return fmt.Errorf("failed to read CA cert: %w", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertPEM) {
		return fmt.Errorf("failed to append CA cert to pool")
	}

	// expected server fingerprint (optional) - file may not exist in some installs
	var expectedServerFP string
	if b, err := os.ReadFile(configDir + "/server_fingerprint_sha256"); err == nil {
		expectedServerFP = strings.TrimSpace(string(b))
	}

	fmt.Println("expectedServerFP: ", expectedServerFP)
	// Read the server fingerprint file used in VerifyPeerCertificate, if available
	// We'll build tls.Config dynamically in each attempt so VerifyPeerCertificate closure can capture expectedServerFP.

	baseURL := os.Getenv("WS_BASE_URL")
	url := strings.TrimRight(baseURL, "/") + "/agent/connect"

	// We'll attempt to connect in a loop until ctx is done
	attempt := 0
	backoff := initialBackoff

	for {
		// Check cancel
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Build TLS config (captures expectedServerFP)
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{clientCert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS13,
		}

		if expectedServerFP != "" {
			// use fingerprint pinning on server cert
			tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
				if len(rawCerts) == 0 {
					return fmt.Errorf("no server cert presented")
				}
				sum := sha256.Sum256(rawCerts[0])
				// fmt.Printf(" rawCerts %+v \n", string(rawCerts[0]))

				if hex.EncodeToString(sum[:]) != expectedServerFP {
					return fmt.Errorf("---server certificate fingerprint mismatch")
				}
				return nil
			}
		}

		dialer := websocket.Dialer{
			TLSClientConfig: tlsConfig,
		}

		log.Printf("Attempting WebSocket dial to %s (attempt=%d)", url, attempt+1)
		conn, resp, err := dialer.Dial(url, nil)
		if err != nil {
			// log http response body if present
			if resp != nil {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				log.Printf("dial error (status=%v): %v - body: %s", resp.Status, err, string(body))
			} else {
				log.Printf("dial error: %v", err)
			}

			// backoff with jitter
			if ctx.Err() != nil {
				return ctx.Err()
			}
			attempt++
			if maxAttempts > 0 && attempt >= maxAttempts {
				return fmt.Errorf("max attempts reached: %v", err)
			}
			// sleep with jitter
			jitter := time.Duration(rand.Intn(500)) * time.Millisecond
			sleep := backoff + jitter
			if sleep > maxBackoff {
				sleep = maxBackoff
			}
			log.Printf("Reconnect sleeping %v before next attempt", sleep)
			time.Sleep(sleep)
			// increase backoff (exponential), cap it
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// Successfully connected - reset attempt/backoff counters
		attempt = 0
		backoff = initialBackoff

		// Set read limits / handlers
		conn.SetReadLimit(readLimit)
		conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})

		log.Println("***WebSocket connection established with mutual TLS!")

		// Send optional hello
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"hello---agent_id":"`+utils.GetAgentID()+`"}`))

		// load signature secret (trimmed)
		bs, err := os.ReadFile(configDir + "/signature_secret")
		if err != nil {
			_ = conn.Close()
			return fmt.Errorf("failed to load signature secret: %w", err)
		}
		signatureSecret := bytes.TrimSpace(bs)

		// start read loop and heartbeat writer
		errCh := make(chan error, 1)
		// read loop
		go func() {
			for {
				_, msg, rerr := conn.ReadMessage()
				if rerr != nil {
					errCh <- rerr
					return
				}
				// handle messages (instructions or anything else)
				if len(msg) > 0 {
					log.Printf("Received from server: %s", string(msg))

					handleIncomingTask(msg)
				}
			}
		}()

		// heartbeat ticker + write loop
		ticker := time.NewTicker(5 * time.Second)
		// Also maintain ping ticker to keep connection alive
		pingTicker := time.NewTicker(pingPeriod)

		connected := true

		for connected {
			select {
			case <-ctx.Done():
				log.Println("context canceled, closing connection")
				ticker.Stop()
				pingTicker.Stop()
				_ = conn.Close()
				return ctx.Err()
			case rerr := <-errCh:
				// read loop encountered an error (connection closed/unexpected)
				log.Printf("connection read error: %v", rerr)
				ticker.Stop()
				pingTicker.Stop()
				_ = conn.Close()
				connected = false
			case <-pingTicker.C:
				// send ping
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("ping write error: %v", err)
					ticker.Stop()
					pingTicker.Stop()
					_ = conn.Close()
					connected = false
				}
			case <-ticker.C:
				// send signed heartbeat
				msg := utils.PrepareHeartbeatMessage(signatureSecret)
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					log.Printf("heartbeat write error: %v", err)
					ticker.Stop()
					pingTicker.Stop()
					_ = conn.Close()
					connected = false
				} else {
					log.Println("Heartbeat sent")
				}
			}
		}

		log.Println("connection lost; will attempt to reconnect")
		// small backoff before next attempt
		time.Sleep(500 * time.Millisecond)
	}
}

func ConnectWithAssistant(ctx context.Context) {

	go func() {
		if err := ConnectAndMaintain(ctx); err != nil {
			log.Printf("connect loop exited: %v", err)
		}
	}()
	// cancel() later to stop
}
func handleIncomingTask(raw []byte) []byte {
	// raw is the JSON payload the assistant sent
	resp, err := runner.ExecuteSignedTaskJSON(raw)
	if err != nil {
		log.Printf("task exec error: %v", err)
		// Return a structured error to assistant
		errObj := map[string]string{"type": "task_result", "status": "error", "error": err.Error()}
		b, _ := json.Marshal(errObj)
		return b
	}

	// os.MkdirAll("logs", os.ModePerm)
	// logFile, err := os.OpenFile("logs/ultaai/agent.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	// if err != nil {
	// 	log.Fatalf("Failed to create log file: %v", err)
	// }
	// log.SetOutput(logFile)

	// Wrap it so assistant can tell it's a result
	return resp // already contains TaskResult JSON; you can wrap with {"type":"task_result", ...} if you want

}
