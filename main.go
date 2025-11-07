package main

import (
	"log"

	"github.com/pmoieni/project-racer-server/internal/net"
	"github.com/pmoieni/project-racer-server/internal/services/telemetry"
)

func main() {
	telemtryService, err := telemetry.New()
	if err != nil {
		log.Fatal(err)
	}

	srv := net.NewServer(&net.ServerFlags{
		Host: "localhost",
		Port: 1234,
	}, telemtryService)

	if err := srv.Run("", ""); err != nil {
		log.Fatalf("Could not start the server: %v", err)
	}
}
