package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	_ "embed"

	"github.com/alecthomas/assert/v2"
	"github.com/gorilla/websocket"
	"oss.acmcsuf.com/qg/backend/internal/hc"
	"oss.acmcsuf.com/qg/backend/internal/west"
	"oss.acmcsuf.com/qg/backend/qg"
	"oss.acmcsuf.com/qg/backend/qg/stores/sqlite"
)

var jeopardyGameData = qg.JeopardyGameData{
	Categories: []qg.JeopardyCategory{
		{
			Name: "Lorem Ipsum 1",
			Questions: []qg.JeopardyQuestion{
				{Question: "1"},
				{Question: "2"},
				{Question: "3"},
			},
		},
		{
			Name: "Lorem Ipsum 2",
			Questions: []qg.JeopardyQuestion{
				{Question: "4"},
				{Question: "5"},
				{Question: "6"},
			},
		},
	},
}

func TestJeopardyWebsocket(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	store, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatal("failed to open SQLite DB:", err)
	}

	handler := newHandler(store)
	t.Cleanup(func() { handler.Close() })

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client := hc.NewClient(srv.URL, srv.Client())
	client.Timeout = 2 * time.Second

	var gameID string

	t.Run("new_game", func(t *testing.T) {
		r, err := hc.POST[qg.ResponseNewGame](ctx, client, "/game",
			qg.RequestNewGame{
				AdminPassword: "admin",
				Data: qg.GameData{
					Value: qg.GameDataJeopardy{Data: jeopardyGameData},
				},
			},
		)
		if err != nil {
			t.Fatal("failed to create new game:", err)
		}

		gameID = r.GameID
	})

	t.Run("get_game", func(t *testing.T) {
		r, err := hc.GET[qg.ResponseGetGame](ctx, client, "/game/"+gameID, nil)
		if err != nil {
			t.Fatal("failed to get game:", err)
		}

		assert.Equal(t, r.GameType, qg.GameTypeJeopardy)
	})

	type gameSequencer struct {
		who string
		act func(t *testing.T, ctx context.Context, ws *west.WebsocketTest)
	}

	sequences := []gameSequencer{
		{
			who: "player 1",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				sendCommand(ctx, t, ws, qg.CommandJoinGame{
					GameID:     gameID,
					PlayerName: "Player 1",
				})

				game := expectEvent[qg.EventJoinedGame](ctx, t, ws)

				gameInfo, ok := game.GameInfo.Value.(qg.GameInfoJeopardy)
				assert.True(t, ok, fmt.Sprintf("unexpected game info type: %T", game.GameInfo))
				assert.Equal(t, gameInfo.Data, qg.JeopardyGameInfo{
					Categories: []string{
						"Lorem Ipsum 1",
						"Lorem Ipsum 2",
					},
					NumQuestions:    3,
					ScoreMultiplier: 100,
				})

				player := expectEvent[qg.EventPlayerJoined](ctx, t, ws)
				assert.Equal(t, player.PlayerName, "Player 1")
			},
		},
		{
			who: "admin",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				sendCommand(ctx, t, ws, qg.CommandJoinGame{
					GameID:        gameID,
					PlayerName:    "Admin",
					AdminPassword: p("admin"),
				})

				game := expectEvent[qg.EventJoinedGame](ctx, t, ws)
				assert.True(t, game.IsAdmin)
				assert.Equal(t,
					game.GameData,
					&qg.GameData{Value: qg.GameDataJeopardy{Data: jeopardyGameData}},
				)

				gameInfo := game.GameInfo.Value.(qg.GameInfoJeopardy).Data
				assert.Equal(t, gameInfo, qg.JeopardyGameInfo{
					Categories: []string{
						"Lorem Ipsum 1",
						"Lorem Ipsum 2",
					},
					NumQuestions:    3,
					ScoreMultiplier: 100,
				})

				player1 := expectEvent[qg.EventPlayerJoined](ctx, t, ws)
				player2 := expectEvent[qg.EventPlayerJoined](ctx, t, ws)
				assertSetEqual(t,
					[]string{player1.PlayerName, player2.PlayerName},
					[]string{"Player 1", "Admin"},
				)
			},
		},
		{
			who: "player 1",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				player := expectEvent[qg.EventPlayerJoined](ctx, t, ws)
				assert.Equal(t, player.PlayerName, "Admin")
			},
		},
		{
			who: "admin",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				sendCommand(ctx, t, ws, qg.CommandBeginGame{})

				expectEvent[qg.EventGameStarted](ctx, t, ws)

				turn := expectEvent[qg.EventJeopardyTurnEnded](ctx, t, ws)
				assert.Equal(t, turn.Chooser, "Player 1")
				assert.Equal(t, turn.Answered, qg.JeopardyAnsweredQuestions{})
			},
		},
		{
			who: "player 1",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				expectEvent[qg.EventGameStarted](ctx, t, ws)

				turn := expectEvent[qg.EventJeopardyTurnEnded](ctx, t, ws)
				assert.Equal(t, turn.Chooser, "Player 1")
				assert.Equal(t, turn.Answered, qg.JeopardyAnsweredQuestions{})

				sendCommand(ctx, t, ws, qg.CommandJeopardyChooseQuestion{
					Category: 0,
					Question: 0,
				})

				question := expectEvent[qg.EventJeopardyBeginQuestion](ctx, t, ws)
				assert.Equal(t, question.Category, 0)
				assert.Equal(t, question.Question, "1")
				assert.Equal(t, question.Points, 100)
			},
		},
		{
			who: "admin",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				question := expectEvent[qg.EventJeopardyBeginQuestion](ctx, t, ws)
				assert.Equal(t, question.Category, 0)
				assert.Equal(t, question.Question, "1")
				assert.Equal(t, question.Points, 100)
			},
		},
		{
			who: "player 1",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				sendCommand(ctx, t, ws, qg.CommandJeopardyPressButton{})

				press := expectEvent[qg.EventJeopardyButtonPressed](ctx, t, ws)
				assert.Equal(t, press.PlayerName, "Player 1")
			},
		},
		{
			who: "admin",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				press := expectEvent[qg.EventJeopardyButtonPressed](ctx, t, ws)
				assert.Equal(t, press.PlayerName, "Player 1")
			},
		},
		{
			who: "admin",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				sendCommand(ctx, t, ws, qg.CommandJeopardyPlayerJudgment{
					Correct: true,
				})

				turn := expectEvent[qg.EventJeopardyTurnEnded](ctx, t, ws)
				assert.Equal(t, turn.Chooser, "Player 1")
				assert.Equal(t, turn.Answered, qg.JeopardyAnsweredQuestions{
					{Category: 0, Question: 0, Player: "Player 1"},
				})
			},
		},
		{
			who: "player 1",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				turn := expectEvent[qg.EventJeopardyTurnEnded](ctx, t, ws)
				assert.Equal(t, turn.Chooser, "Player 1")
				assert.Equal(t, turn.Answered, qg.JeopardyAnsweredQuestions{
					{Category: 0, Question: 0, Player: "Player 1"},
				})

				sendCommand(ctx, t, ws, qg.CommandJeopardyChooseQuestion{
					Category: 1,
					Question: 1,
				})

				question := expectEvent[qg.EventJeopardyBeginQuestion](ctx, t, ws)
				assert.Equal(t, question.Category, 1)
				assert.Equal(t, question.Question, "5")
				assert.Equal(t, question.Points, 200)
			},
		},
		{
			who: "admin",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				question := expectEvent[qg.EventJeopardyBeginQuestion](ctx, t, ws)
				assert.Equal(t, question.Category, 1)
				assert.Equal(t, question.Question, "5")
				assert.Equal(t, question.Points, 200)
			},
		},
		{
			who: "player 1",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				sendCommand(ctx, t, ws, qg.CommandJeopardyPressButton{})

				press := expectEvent[qg.EventJeopardyButtonPressed](ctx, t, ws)
				assert.Equal(t, press.PlayerName, "Player 1")
			},
		},
		{
			who: "admin",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				press := expectEvent[qg.EventJeopardyButtonPressed](ctx, t, ws)
				assert.Equal(t, press.PlayerName, "Player 1")

				sendCommand(ctx, t, ws, qg.CommandJeopardyPlayerJudgment{
					Correct: false,
				})

				turn := expectEvent[qg.EventJeopardyTurnEnded](ctx, t, ws)
				assert.Equal(t, turn.Chooser, "Player 1")
				assert.Equal(t, turn.Answered, qg.JeopardyAnsweredQuestions{
					{Category: 0, Question: 0, Player: "Player 1"},
				})
			},
		},
		{
			who: "player 1",
			act: func(t *testing.T, ctx context.Context, ws *west.WebsocketTest) {
				turn := expectEvent[qg.EventJeopardyTurnEnded](ctx, t, ws)
				assert.Equal(t, turn.Chooser, "Player 1")
				assert.Equal(t, turn.Answered, qg.JeopardyAnsweredQuestions{
					{Category: 0, Question: 0, Player: "Player 1"},
				})
			},
		},
	}

	t.Run("play_game", func(t *testing.T) {
		var wg sync.WaitGroup
		t.Cleanup(func() { wg.Wait() })

		ctx, cancel := context.WithCancelCause(ctx)
		t.Cleanup(func() {
			if t.Failed() {
				cancel(errors.New("test failed"))
				return
			}

			if err := context.Cause(ctx); err != nil {
				t.Error("context was cancelled:", err)
			}

			cancel(nil)
		})

		sequencers := make(map[string]*west.WebsocketTest)
		for _, sequence := range sequences {
			sequencers[sequence.who] = nil
		}

		for who := range sequencers {
			who := who

			ws, err := west.NewTestWebsocket(srv.URL+"/ws", &websocket.Dialer{
				HandshakeTimeout: 2 * time.Second,
			})
			if err != nil {
				t.Fatal("failed to create websocket:", err)
			}

			ws.LogSent = func(msg json.RawMessage) {
				t.Helper()
				t.Log("sent", who+":", string(msg))
			}

			ws.LogReceived = func(msg json.RawMessage) {
				t.Helper()
				t.Log("recv", who+":", string(msg))
			}

			wg.Add(1)
			go func() {
				cancel(ws.Start(ctx))
				wg.Done()
			}()

			sequencers[who] = ws
		}

		for i, sequence := range sequences {
			t.Logf("sequence %d: %s", i, sequence.who)
			sequencer := sequencers[sequence.who]
			sequence.act(t, ctx, sequencer)
		}
	})
}

func assertSetEqual[T comparable](t *testing.T, a, b []T) {
	t.Helper()

	var fail bool
aLoop:
	for _, av := range a {
		for _, bv := range b {
			if av == bv {
				continue aLoop
			}
		}

		t.Errorf("%v in a is not in b", av)
		fail = true
	}

bLoop:
	for _, bv := range b {
		for _, av := range a {
			if av == bv {
				continue bLoop
			}
		}

		t.Errorf("%v in b is not in a", bv)
		fail = true
	}

	if fail {
		t.Fatalf("%v != %v", a, b)
	}
}

func sendCommand(ctx context.Context, t *testing.T, ws *west.WebsocketTest, cmd qg.ICommand) {
	t.Helper()
	err := ws.Send(ctx, cmd)
	must(t, err)
}

func expectEvent[T qg.IEvent](ctx context.Context, t *testing.T, w *west.WebsocketTest) T {
	t.Helper()

	event, err := west.Expect[T](ctx, w)
	assert.NoError(t, err)
	assert.True(t, event != nil)

	return *event
}

func p[T any](v T) *T {
	return &v
}

func must(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func marshal(v any) io.Reader {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return bytes.NewReader(b)
}
