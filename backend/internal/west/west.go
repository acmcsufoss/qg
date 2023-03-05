package west

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

// WebsocketTest is a test helper for Websockets.
type WebsocketTest struct {
	dialer *websocket.Dialer
	wsURL  *url.URL

	closedq chan struct{}
	expectq chan expectation
	sendq   chan any

	ExpectTimeout time.Duration
	LogReceived   func(json.RawMessage)
	LogSent       func(json.RawMessage)
}

// NewTestWebsocket creates a new WebsocketTest.
func NewTestWebsocket(addr string, dialer *websocket.Dialer) (*WebsocketTest, error) {
	if dialer == nil {
		dialer = websocket.DefaultDialer
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, errors.Wrap(err, "parsing URL")
	}

	if strings.HasPrefix(u.Scheme, "http") {
		switch u.Scheme {
		case "http":
			u.Scheme = "ws"
		case "https":
			u.Scheme = "wss"
		default:
			return nil, errors.Errorf("invalid scheme %q", u.Scheme)
		}
	}

	return &WebsocketTest{
		dialer:  dialer,
		wsURL:   u,
		closedq: make(chan struct{}),
		expectq: make(chan expectation),
		sendq:   make(chan any),

		ExpectTimeout: 5 * time.Second,
	}, nil
}

// Start starts the WebsocketTest event loop and blocks until the context is
// canceled.
func (w *WebsocketTest) Start(ctx context.Context) error {
	return w.eventLoop(ctx)
}

func (w *WebsocketTest) eventLoop(ctx context.Context) (err error) {
	var wg sync.WaitGroup
	defer wg.Wait()

	conn, _, err := w.dialer.DialContext(ctx, w.wsURL.String(), nil)
	if err != nil {
		return errors.Wrap(err, "dialing websocket")
	}
	defer conn.Close()

	ctx, cancel := context.WithCancelCause(ctx)
	defer func() { err = context.Cause(ctx) }()
	defer cancel(nil)

	defer close(w.closedq)

	recvch := make(chan json.RawMessage)
	wg.Add(1)
	go func(recvch chan<- json.RawMessage) {
		defer wg.Done()

		for {
			_, b, err := conn.ReadMessage()
			if err != nil {
				cancel(err)
				return
			}

			if w.LogReceived != nil {
				w.LogReceived(b)
			}

			select {
			case recvch <- b:
				// ok
			case <-ctx.Done():
				return
			}
		}
	}(recvch)

	sendch := w.sendq
	wg.Add(1)
	go func(sendch <-chan any) {
		defer wg.Done()

		for {
			select {
			case v := <-sendch:
				b, err := json.Marshal(v)
				if err != nil {
					cancel(errors.Wrap(err, "marshaling message"))
					return
				}

				if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
					cancel(err)
					return
				}

				if w.LogSent != nil {
					w.LogSent(b)
				}

			case <-ctx.Done():
				return
			}
		}
	}(sendch)

	recvq := make([]json.RawMessage, 0, 10)
	expectq := make([]expectation, 0, 10)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case expect := <-w.expectq:
			expectq = append(expectq, expect)

		case b := <-recvch:
			recvq = append(recvq, b)
		}

		// For each expectations, try to match it with a received message.
	searchExpectations:
		for i := 0; i < len(expectq); i++ {
			expect := expectq[i]

			for j := 0; j < len(recvq); j++ {
				recv := recvq[j]

				v, err := expect.unmarshalRecv(recv)
				if err == nil {
					// Found! Dispatch and remove.
					expect.res(v.Interface())
					expectq = append(expectq[:i], expectq[i+1:]...)
					recvq = append(recvq[:j], recvq[j+1:]...)
					continue searchExpectations
				}
			}
		}
	}
}

// Send sends directly to the websocket. Use this only for initial handshakes.
func (w *WebsocketTest) Send(ctx context.Context, v any) error {
	select {
	case w.sendq <- v:
		return nil
	case <-w.closedq:
		return net.ErrClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

type expectation struct {
	rtyp reflect.Type
	resq chan any
}

func (e expectation) expectType() reflect.Type {
	return e.rtyp
}

func (e expectation) res(v any) {
	e.resq <- v
}

// unmarshalRecv unmarshals the received message into recv.
func (e expectation) unmarshalRecv(b []byte) (reflect.Value, error) {
	unmarshalType := e.rtyp
	if unmarshalType.Kind() != reflect.Struct {
		return reflect.Value{}, errors.Errorf("unsupported type %v", unmarshalType)
	}

	// Allocate a value of the expected type.
	v := reflect.New(unmarshalType)

	if err := json.Unmarshal(b, v.Interface()); err != nil {
		return reflect.Value{}, errors.Wrap(err, "unmarshaling message")
	}

	v = v.Elem()

	// Ensure that all fields are not zero.

	return v, nil
}

// Expect expects a message on the websocket and a certain response given that
// message.
func Expect[T any](ctx context.Context, w *WebsocketTest) (*T, error) {
	ctx, cancel := context.WithTimeout(ctx, w.ExpectTimeout)
	defer cancel()

	var zero T

	expect := expectation{
		rtyp: reflect.TypeOf(zero),
		resq: make(chan any, 1),
	}

	select {
	case w.expectq <- expect:
		// ok
	case <-w.closedq:
		return nil, net.ErrClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case val := <-expect.resq:
		switch v := val.(type) {
		case error:
			return nil, v
		case T:
			return &v, nil
		default:
			return nil, errors.Errorf("unexpected type %T", v)
		}
	case <-w.closedq:
		return nil, net.ErrClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
