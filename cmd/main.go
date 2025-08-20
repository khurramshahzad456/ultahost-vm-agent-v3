package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"ultahost-ai-gateway/internal/config"
	"ultahost-ai-gateway/internal/server"
	"ultahost-ai-gateway/internal/websocket"

	ws "github.com/gorilla/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables
	config.LoadConfig()

	// Initialize server
	s := server.NewServer()

	// Register all routes
	server.RegisterRoutes(s.Engine)

	//server starting
	go wsTlsInit()

	log.Printf(" Server starting on port %s...\n", config.AppConfig.Port)
	if err := s.Start(); err != nil {
		log.Fatalf(" Server failed to start: %v", err)
	}

}

// Upgrader with CheckOrigin allowing all origins (for demo; restrict in production)
var upgrader = ws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsTlsInit() {
	// connectedVPS := make(map[string]*websocket.Conn)
	// Load CA cert to verify client certs (mutual TLS)
	caCertPEM, err := ioutil.ReadFile("./certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCertPEM) {
		log.Fatal("Failed to add CA cert to pool")
	}

	// Setup TLS config for server
	tlsConfig := &tls.Config{
		ClientAuth: tls.RequireAndVerifyClientCert, // Require client cert and verify
		ClientCAs:  caCertPool,                     // Client certs must be signed by this CA
		MinVersion: tls.VersionTLS13,
	}

	// Setup Gin router
	r := gin.Default()

	r.GET("/agent/connect", websocket.HandleAgentWebSocket)

	// Create HTTPS server with TLS config
	server := &http.Server{
		Addr:      ":8443",
		Handler:   r,
		TLSConfig: tlsConfig,
	}

	log.Println("Starting TLS WebSocket server on https://localhost:8443 ...")
	err = server.ListenAndServeTLS("./certs/server.crt", "./certs/server.key")
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
