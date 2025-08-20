// internal/websocket/agent_websocket.go
package websocket

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"ultahost-ai-gateway/internal/utils"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// restrict if you know allowed origins; keep true for LAN/dev
		return true
	},
}

type AgentConn struct {
	Conn                 *ws.Conn
	IdentityToken        string
	LastHeartbeatCounter uint64
	LastSeen             time.Time
	mu                   sync.Mutex
}

var (
	connectedMtx sync.RWMutex
	ConnectedVPS = make(map[string]*AgentConn) // identityToken -> AgentConn
)

const (
	readTimeout  = 60 * time.Second
	writeTimeout = 15 * time.Second
	pongWait     = 70 * time.Second
	pingPeriod   = 30 * time.Second
)

func HandleAgentWebSocket(c *gin.Context) {
	if c.Request.TLS == nil || len(c.Request.TLS.PeerCertificates) == 0 {
		c.String(http.StatusUnauthorized, "Client certificate required")
		return
	}

	clientCert := c.Request.TLS.PeerCertificates[0]
	cn := clientCert.Subject.CommonName

	// Upgrade to websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Load agent keys
	keyInfo, exist := utils.GetAgentKeys(cn)
	if !exist {
		conn.WriteMessage(ws.TextMessage, []byte("agent not enrolled"))
		conn.Close()
		return
	}

	// Compute fingerprint of presented cert and compare
	presentedFP := sha256.Sum256(clientCert.Raw)
	fmt.Println(" --------------------------- keyInfo.FingerprintSHA256:  ", keyInfo.FingerprintSHA256)
	fmt.Println(" --------------------------- clientCert.Raw:  ", hex.EncodeToString(presentedFP[:]))

	// if hex.EncodeToString(clientCert.Raw[:]) != keyInfo.FingerprintSHA256 {
	if hex.EncodeToString(presentedFP[:]) != keyInfo.FingerprintSHA256 {

		conn.WriteMessage(ws.TextMessage, []byte("certificate fingerprint mismatch"))
		conn.Close()

		return
	}

	// Build AgentConn and register
	agentConn := &AgentConn{Conn: conn, IdentityToken: keyInfo.IdentityToken, LastSeen: time.Now()}
	connectedMtx.Lock()
	ConnectedVPS[keyInfo.IdentityToken] = agentConn
	connectedMtx.Unlock()

	// Setup ping/pong and deadlines
	conn.SetReadLimit(1024 * 1024)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		agentConn.mu.Lock()
		agentConn.LastSeen = time.Now()
		agentConn.mu.Unlock()
		return nil
	})

	// Read loop
	go handleAgentReadLoop(agentConn, keyInfo)
	// Ping loop (server -> client)
	go handleAgentPingLoop(agentConn)

	log.Printf("Agent connected: CN=%s, IdentityToken=%s", cn, keyInfo.IdentityToken)
}

func handleAgentPingLoop(a *AgentConn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		a.Conn.Close()
	}()
	for range ticker.C {
		a.mu.Lock()
		a.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		if err := a.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
			a.mu.Unlock()
			return
		}
		a.mu.Unlock()
	}
}

// func handleAgentReadLoop(a *AgentConn, keyInfo utils.AgentKeys) {
// 	defer func() {
// 		connectedMtx.Lock()
// 		delete(ConnectedVPS, a.IdentityToken)
// 		connectedMtx.Unlock()
// 		a.Conn.Close()
// 	}()

// 	for {
// 		mt, msg, err := a.Conn.ReadMessage()
// 		if err != nil {
// 			log.Printf("read error: %v", err)
// 			return
// 		}
// 		a.mu.Lock()
// 		a.LastSeen = time.Now()
// 		a.mu.Unlock()

// 		if mt == ws.TextMessage {
// 			// attempt parse heartbeat or command ack
// 			var generic map[string]interface{}
// 			if err := json.Unmarshal(msg, &generic); err == nil {
// 				// log.Printf("--------------got msg: %+v \n", generic)
// 				if t, ok := generic["type"].(string); ok && t == "heartbeat" {
// 					if err := verifyHeartbeat(msg, keyInfo); err != nil {
// 						log.Printf("heartbeat verification failed: %v", err)
// 						// Optionally close connection
// 						return
// 					}
// 					// heartbeat valid, update state (counter, last seen)
// 					log.Printf("--------****************************------got msg: %+v \n", generic)

// 					continue
// 				}
// 			}
// 			// otherwise handle other messages
// 			log.Printf("msg from %s: %s", a.IdentityToken, string(msg))
// 		}
// 	}
// }

func handleAgentReadLoop(a *AgentConn, keyInfo utils.AgentKeys) {
	defer func() {
		connectedMtx.Lock()
		delete(ConnectedVPS, a.IdentityToken)
		connectedMtx.Unlock()

		// fail all pending tasks intended for this agent so waiters don't hang
		failPendingForAgent(keyInfo.IdentityToken, "connection closed")
		a.Conn.Close()
	}()

	for {
		mt, msg, err := a.Conn.ReadMessage()
		if err != nil {
			log.Printf("read error: %v", err)
			return
		}
		a.mu.Lock()
		a.LastSeen = time.Now()
		a.mu.Unlock()

		if mt == ws.TextMessage {
			// attempt parse heartbeat or command ack
			var generic map[string]interface{}
			if err := json.Unmarshal(msg, &generic); err == nil {
				if t, ok := generic["type"].(string); ok {
					switch t {
					case "heartbeat":
						// existing heartbeat verification
						if err := verifyHeartbeat(msg, keyInfo); err != nil {
							log.Printf("heartbeat verification failed: %v", err)
							return
						}
						log.Println("Heartbeat msg: ", string(msg))

						continue

					case "task_result":
						// parse into TaskResult and resolve pending waiter
						var tr TaskResult
						if err := json.Unmarshal(msg, &tr); err != nil {
							log.Printf("invalid task_result format from %s: %v", keyInfo.IdentityToken, err)
							continue
						}
						if resolved := resolvePending(tr.TaskID, tr); resolved {
							log.Printf("resolved pending task %s for agent %s", tr.TaskID, keyInfo.IdentityToken)
						} else {
							log.Printf("received task_result for unknown task_id %s (agent %s)", tr.TaskID, keyInfo.IdentityToken)
						}
						continue

					default:
						// unknown message type; fallthrough to general log
					}
				}
			}
			// otherwise handle other messages
			log.Printf("msg from %s: %s", a.IdentityToken, string(msg))
		}
	}
}

// verifyHeartbeat parses heartbeat JSON, checks signature using keyInfo.SignatureSecret
func verifyHeartbeat(msg []byte, keyInfo utils.AgentKeys) error {
	type hb struct {
		Type      string `json:"type"`
		Version   int    `json:"version"`
		AgentID   string `json:"agent_id"`
		Counter   uint64 `json:"counter"`
		Nonce     string `json:"nonce"`
		Timestamp string `json:"timestamp"`
		Signature string `json:"signature"`
	}
	var h hb
	if err := json.Unmarshal(msg, &h); err != nil {
		return err
	}

	// Timestamp skew check (±5m by default)
	maxSkew := 5 * time.Minute
	ht, err := time.Parse(time.RFC3339Nano, h.Timestamp)
	if err != nil {
		// Try RFC3339 fallback
		ht, err = time.Parse(time.RFC3339, h.Timestamp)
		if err != nil {
			return fmt.Errorf("invalid heartbeat timestamp: %w", err)
		}
	}
	delta := time.Since(ht.UTC())
	if delta < 0 {
		delta = -delta
	}
	if delta > maxSkew {
		return errors.New("heartbeat timestamp outside allowed skew")
	}

	// Rebuild canonical string
	canon := fmt.Sprintf("%s|%s|%d|%s|%s", h.Version, h.AgentID, h.Counter, h.Nonce, h.Timestamp)
	// verify HMAC
	expected := utils.HMACSHA256Base64([]byte(keyInfo.SignatureSecret), canon)
	if expected != h.Signature {
		return errors.New("invalid signature")
	}

	// TODO: store/compare counter to prevent replay. For demo, accept increasing counter check:
	// You can persist lastCounter per agent (DB) in production. Here we keep in-memory in ConnectedVPS.
	connectedMtx.RLock()
	aConn, ok := ConnectedVPS[keyInfo.IdentityToken]
	connectedMtx.RUnlock()
	if ok {
		if h.Counter <= aConn.LastHeartbeatCounter {
			return errors.New("replay or old counter")
		}
		aConn.LastHeartbeatCounter = h.Counter
	}
	return nil
}
