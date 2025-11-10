package telemetry

import (
	"fmt"
	"net/http"

	"github.com/pmoieni/project-racer-server/internal/lib"
	"github.com/pmoieni/project-racer-server/internal/net"
	"github.com/pmoieni/project-racer-server/internal/net/websocket"
)

var _ net.Service = (*TelemetryService)(nil)

type TelemetryService struct {
	*http.ServeMux

	hub *websocket.Hub
	log *lib.Logger
}

func New() (*TelemetryService, error) {
	s := &TelemetryService{
		ServeMux: http.NewServeMux(),
		hub:      websocket.NewHub(),
		log:      lib.NewLogger("telemetry"),
	}
	s.setupControllers()

	return s, nil
}

func (s *TelemetryService) MountPath() string {
	return "telemetry"
}

func (s *TelemetryService) setupControllers() {
	s.HandleFunc("GET /ws", handleConn(s.hub))
}

func handleConn(hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: do some checks here
		hub.ServeHTTP(w, r)
	}
}
