package server

import (
	"context"
	"io"
	"net/http"

	"etok.codes/qg/backend/internal/hrt"
	"etok.codes/qg/backend/qg"
	"etok.codes/qg/backend/qg/games/jeopardy"
	"etok.codes/qg/backend/server/ws"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

// Storer is a storage interface for the server.
type Storer interface {
	qg.GameStorer
	jeopardy.Storer
}

type handler struct {
	*chi.Mux
	ws  *ws.Handler
	api *apiHandler
}

// HTTPHandlerCloser is a handler that can be closed.
type HTTPHandlerCloser interface {
	http.Handler
	io.Closer
}

// NewHandler creates a new Handler.
func NewHandler(storer Storer) HTTPHandlerCloser {
	h := &handler{
		Mux: chi.NewMux(),
		ws:  ws.NewHandler(),
		api: &apiHandler{},
	}

	h.Use(hrt.Use(hrt.Opts{
		Encoder:     hrt.EncoderWithValidator(hrt.JSONEncoder),
		ErrorWriter: hrt.WriteErrorFunc(writeError),
	}))

	h.Mount("/ws", h.ws)

	h.Route("/game", func(r chi.Router) {
		r.Get("/", hrt.Wrap(h.api.getGame))
		r.Post("/", hrt.Wrap(h.api.postGame))

		r.Get("/jeopardy", hrt.Wrap(h.api.getJeopardy))
	})

	return h
}

func (h *handler) Close() error {
	h.ws.Stop()
	return nil
}

func writeError(w http.ResponseWriter, err error) {
	qg.WriteHTTPError(w, hrt.ErrorHTTPStatus(err, 500), err)
}

type apiHandler struct {
	store Storer
}

func (h *apiHandler) getGame(ctx context.Context, body qg.RequestGetGame) (qg.ResponseGetGame, error) {
	gameType, err := h.store.GameType(ctx, body.GameID)
	if err != nil {
		return qg.ResponseGetGame{}, err
	}

	return qg.ResponseGetGame{GameType: gameType}, nil
}

func (h *apiHandler) postGame(ctx context.Context, body qg.RequestNewGame) (qg.ResponseNewGame, error) {
	gameID, err := h.store.MakeNewGame(ctx, body.Data)
	if err != nil {
		return qg.ResponseNewGame{}, err
	}

	if err := h.store.SetGamePassword(ctx, gameID, body.ModeratorPassword); err != nil {
		return qg.ResponseNewGame{}, errors.Wrap(err, "failed to set game password")
	}

	return qg.ResponseNewGame{
		GameID:   gameID,
		GameType: qg.GameTypeFromData(body.Data),
	}, nil
}

func (h *apiHandler) getJeopardy(ctx context.Context, body qg.RequestGetJeopardyGame) (qg.ResponseGetJeopardyGame, error) {
	data, err := h.store.JeopardyGameData(ctx, body.GameID)
	if err != nil {
		return qg.ResponseGetJeopardyGame{}, err
	}

	info := qg.ConvertJeopardyGameData(data)
	return qg.ResponseGetJeopardyGame{Info: info}, nil
}
