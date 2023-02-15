package ws

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"oss.acmcsuf.com/qg/backend/qg"
)

// DefaultUpgrader is the default upgrader used by the websocket server.
var DefaultUpgrader = websocket.Upgrader{
	EnableCompression: true,
}

// Handler is a handler for websocket connections.
type Handler struct {
	// Upgrader is the websocket upgrader. It is used to upgrade HTTP requests
	// to websocket connections.
	Upgrader websocket.Upgrader

	wg   sync.WaitGroup
	srvs sync.Map
	hfac qg.CommandHandlerFactory
}

// NewHandler creates a new websocket handler.
func NewHandler(hfac qg.CommandHandlerFactory) *Handler {
	return &Handler{
		Upgrader: DefaultUpgrader,
		hfac:     hfac,
	}
}

// Stop stops all servers. The function blocks until all servers have been
// stopped.
func (h *Handler) Stop() {
	h.srvs.Range(func(k, _ any) bool {
		ch := k.(chan qg.Event)
		close(ch)
		return true
	})
	h.wg.Wait()
}

// ServeHTTP implements the http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srvh := serverHandler{root: h}
	srvh.ServeHTTP(w, r)
}
