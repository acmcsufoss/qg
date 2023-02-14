package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"etok.codes/qg/backend/qg"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

func newSendLimiter() *rate.Limiter {
	return rate.NewLimiter(rate.Every(2*time.Second), 16)
}

type serverHandler struct {
	root *Handler
	hfac qg.CommandHandlerFactory
}

func (h serverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.root.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		qg.WriteHTTPError(w, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithCancelCause(r.Context())
	defer cancel(nil)

	ch := make(chan qg.Event, 16)

	h.root.srvs.Store(ch, struct{}{})
	defer h.root.srvs.Delete(ch)

	cmdh := h.hfac.CommandHandler()
	cmdh.Subscribe(ctx, ch)
	defer cmdh.Unsubscribe(ctx, ch)

	var wg sync.WaitGroup
	defer wg.Wait()

	server := &server{
		ws: conn,
	}

	wg.Add(1)
	go func() {
		server.eventLoop(ctx, ch)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		server.commandLoop(ctx, cmdh)
		wg.Done()
	}()
}

// server is a websocket server.
type server struct {
	ws     *websocket.Conn
	cancel context.CancelCauseFunc
}

func (s *server) commandLoop(ctx context.Context, cmdh qg.CommandHandler) {
	rl := newSendLimiter()
	for {
		if err := rl.Wait(ctx); err != nil {
			s.cancel(err)
			return
		}

		var cmd qg.Command
		if err := s.ws.ReadJSON(&cmd); err != nil {
			s.cancel(err)
			return
		}

		if err := cmdh.HandleCommand(ctx, cmd); err != nil {
			s.cancel(err)
			return
		}
	}
}

func (s *server) eventLoop(ctx context.Context, ch <-chan qg.Event) {
	defer s.cancel(s.ws.Close())

	const heartrate = 30 * time.Second

	heartbeat := time.NewTicker(heartrate)
	defer heartbeat.Stop()

	resetDeadline := func() {
		deadline := time.Now().Add(2 * heartrate)
		s.ws.SetReadDeadline(deadline)
		s.ws.SetWriteDeadline(deadline)
	}

	s.ws.SetPongHandler(func(string) error {
		resetDeadline()
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			var code int
			var message string
			if err := context.Cause(ctx); err != ctx.Err() {
				if err == nil {
					code = websocket.CloseNormalClosure
				} else {
					code = websocket.CloseInternalServerErr
					message = err.Error()
				}
			} else {
				code = websocket.CloseGoingAway
				message = ctx.Err().Error()
			}

			s.writeClose(code, message)
			return

		case <-heartbeat.C:
			if err := s.writePing(); err != nil {
				s.cancel(err)
				continue
			}

		case event, ok := <-ch:
			if !ok {
				s.cancel(nil)
				continue
			}

			b, err := json.Marshal(event)
			if err != nil {
				s.cancel(err)
				continue
			}

			if err := s.ws.WriteMessage(websocket.TextMessage, b); err != nil {
				s.cancel(err)
				continue
			}
		}
	}
}

const controlMessageTimeout = 5 * time.Second

func (s *server) writeClose(messageCode int, message string) error {
	return s.ws.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(messageCode, message),
		time.Now().Add(controlMessageTimeout))
}

func (s *server) writePing() error {
	return s.ws.WriteControl(
		websocket.PingMessage,
		nil,
		time.Now().Add(controlMessageTimeout))
}
