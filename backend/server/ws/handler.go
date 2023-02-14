package ws

import (
	"net/http"
	"sync"

	"etok.codes/qg/backend/qg"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
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

	r    chi.Router
	wg   sync.WaitGroup
	srvs sync.Map
}

// NewHandler creates a new websocket handler.
func NewHandler() *Handler {
	return &Handler{
		r: chi.NewRouter(),
	}
}

// Stop stops all servers. The function blocks until all servers have been
// stopped.
func (h *Handler) Stop() {
	h.srvs.Range(func(k, _ any) bool {
		ch := k.(chan<- qg.Event)
		close(ch)
		return true
	})
	h.wg.Wait()
}

// Mount mounts a new command handler on the given path.
func (h *Handler) Mount(route string, hfac qg.CommandHandlerFactory) {
	h.r.Mount(route, serverHandler{
		root: h,
		hfac: hfac,
	})
}

// ServeHTTP implements the http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}
