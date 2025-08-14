package main

import (
<<<<<<< Updated upstream
	"fmt"
	"log"
	"os"
=======
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
>>>>>>> Stashed changes

	"ultahost-agent/internal/runner"
)

func main() {
	os.MkdirAll("logs", os.ModePerm)
	logFile, err := os.OpenFile("logs/ultaai/agent.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	log.SetOutput(logFile)
	fmt.Println(" log for binary excuted successfully  ")

<<<<<<< Updated upstream
	scriptPath := "scripts/test_file.sh"
	output, err := runner.ExecuteScript(scriptPath)
=======
	// Assume token is passed as an env var or CLI arg; here env var for demo
	installToken := os.Getenv("INSTALL_TOKEN")
	if installToken == "" {
		log.Fatal("INSTALL_TOKEN not provided")
	}

	// Register agent using install token
	vpsId := "483" //twmporRILY Hrdcoded because Ultahost backend or Nest js API is not completed yet
	err = agent.RegisterAgent(installToken, vpsId)
>>>>>>> Stashed changes
	if err != nil {
		log.Printf(" Script execution failed: %v", err)
	} else {
		log.Printf(" Script executed successfully:\n%s", output)
	}
<<<<<<< Updated upstream
=======
	log.Println("Agent registered successfully")

	time.Sleep(5 * time.Second)
	// Connect to backend WebSocket and start heartbeat
	// err = agent.ConnectAndHeartbeat()
	// if err != nil {
	// 	log.Fatalf("WebSocket connection failed: %v", err)
	// }
	ctx, cancel := context.WithCancel(context.Background())
	agent.ConnectWithAssistant(ctx)

	// Wait for Ctrl+C or kill signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down gracefully...")
	cancel()
	time.Sleep(1 * time.Second) // let goroutines exit
>>>>>>> Stashed changes
}
