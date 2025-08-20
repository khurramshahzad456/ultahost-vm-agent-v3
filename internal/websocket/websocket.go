package websocket

import (
	"fmt"
	"time"
	"ultahost-ai-gateway/internal/utils"

	ws "github.com/gorilla/websocket"
)

// var ConnectedVPS = make(map[string]*websocket.Conn)

// func SendMessage(vpsId string) {
// 	CN := "Agent_" + vpsId
// 	keyInfo, exist := utils.GetAgentKeys(CN)

// 	if exist {
// 		conn, exist := ConnectedVPS[keyInfo.IdentityToken]

// 		if exist {
// 			fmt.Printf("Connected agent found: %+v \n", keyInfo)

// 			conn.WriteMessage(websocket.TextMessage, []byte("df -h"))
// 		} else {
// 			fmt.Println("No connected agent found")
// 		}

// 	}

// }

func SendMessage(vpsId string, payload []byte) error {
	CN := "Agent_" + vpsId
	keyInfo, exist := utils.GetAgentKeys(CN)
	if !exist {
		return fmt.Errorf("no keys")
	}

	connectedMtx.RLock()
	aConn, ok := ConnectedVPS[keyInfo.IdentityToken]
	connectedMtx.RUnlock()
	if !ok {
		return fmt.Errorf("no connected agent")
	}

	aConn.mu.Lock()
	defer aConn.mu.Unlock()
	aConn.Conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	return aConn.Conn.WriteMessage(ws.TextMessage, payload)
}

