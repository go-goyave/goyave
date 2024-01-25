package websocket

import (
	"context"
	stderrors "errors"
	"net/http"
	"strings"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"

	"goyave.dev/goyave/v5/util/errors"
)

// Conn represents a WebSocket connection.
type Conn struct {
	*ws.Conn
	waitClose    chan struct{}
	closeTimeout time.Duration
	closeOnce    sync.Once
}

func newConn(c *ws.Conn, closeTimeout time.Duration) *Conn {
	conn := &Conn{
		Conn:         c,
		waitClose:    make(chan struct{}, 1),
		closeTimeout: closeTimeout,
	}
	c.SetCloseHandler(conn.closeHandler)
	return conn
}

// SetCloseHandshakeTimeout set the timeout used when writing and reading
// close frames during the close handshake.
func (c *Conn) SetCloseHandshakeTimeout(timeout time.Duration) {
	c.closeTimeout = timeout
}

// GetCloseHandshakeTimeout return the timeout used when writing and reading
// close frames during the close handshake.
func (c *Conn) GetCloseHandshakeTimeout() time.Duration {
	return c.closeTimeout
}

func (c *Conn) closeHandler(_ int, _ string) error {
	c.waitClose <- struct{}{}
	return nil
}

// CloseNormal performs the closing handshake as specified by
// RFC 6455 Section 1.4. Sends status code 1000 (normal closure) and
// message "Server closed connection".
//
// This function expects another goroutine to be reading the connection,
// expecting the close frame in response. This waiting can time out. If so,
// Close will just close the connection.
//
// Calling this function multiple times is safe and only the first call will
// write the close frame to the connection.
func (c *Conn) CloseNormal() error {
	return c.Close(ws.CloseNormalClosure, NormalClosureMessage)
}

// CloseWithError performs the closing handshake as specified by
// RFC 6455 Section 1.4 because a server error occurred.
// Sends status code 1011 (internal server error) and
// message "Internal server error".
//
// This function starts another goroutine to read the connection,
// expecting the close frame in response. This waiting can time out. If so,
// Close will just close the connection. Therefore, it is not safe to call
// this function if there is already an active reader.
func (c *Conn) CloseWithError(_ error) error {
	return c.internalClose(ws.CloseInternalServerErr, http.StatusText(http.StatusInternalServerError))
}

// Close performs the closing handshake as specified by RFC 6455 Section 1.4.
//
// This function expects another goroutine to be reading the connection,
// expecting the close frame in response. This waiting can time out. If so,
// Close will just close the connection.
//
// Calling this function multiple times is safe and only the first call will
// write the close frame to the connection.
func (c *Conn) Close(code int, message string) error {
	var err error
	c.closeOnce.Do(func() {
		deadline := time.Now().Add(c.closeTimeout)
		m := ws.FormatCloseMessage(code, message)
		writeErr := c.WriteControl(ws.CloseMessage, m, deadline)
		if writeErr != nil && !stderrors.Is(writeErr, ws.ErrCloseSent) {
			if strings.Contains(writeErr.Error(), "use of closed network connection") {
				err = errors.New(writeErr)
			}
			return
		}

		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		defer cancel()

		select {
		case <-ctx.Done():
		case <-c.waitClose:
			close(c.waitClose)
		}

		err = errors.New(c.Conn.Close())
	})

	return err
}

// internalClose performs the close handshake. Starts a goroutine reading in the connection,
// expecting a close frame response for the close handshake. This function should only be
// used if the server wants to initiate the close handshake.
func (c *Conn) internalClose(code int, message string) error {
	go func() {
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				return
			}
		}
	}()
	return c.Close(code, message)
}
