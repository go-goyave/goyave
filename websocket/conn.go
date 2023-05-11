package websocket

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"goyave.dev/goyave/v4/config"

	ws "github.com/gorilla/websocket"
)

// Conn represents a WebSocket connection.
type Conn struct {
	*ws.Conn
	waitClose chan struct{}
	timeout   time.Duration
	closeOnce sync.Once
}

func newConn(c *ws.Conn) *Conn {
	conn := &Conn{
		Conn:      c,
		waitClose: make(chan struct{}, 1),
		timeout:   timeout,
	}
	c.SetCloseHandler(conn.closeHandler)
	return conn
}

// SetCloseHandshakeTimeout set the timeout used when writing and reading
// close frames during the close handshake.
func (c *Conn) SetCloseHandshakeTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetCloseHandshakeTimeout return the timeout used when writing and reading
// close frames during the close handshake.
func (c *Conn) GetCloseHandshakeTimeout() time.Duration {
	return c.timeout
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
// message "Internal server error". If debug is enabled,
// the message is set to the given error's message.
//
// This function starts another goroutine to read the connection,
// expecting the close frame in response. This waiting can time out. If so,
// Close will just close the connection. Therefore, it is not safe to call
// this function if there is already an active reader.
func (c *Conn) CloseWithError(err error) error {
	message := "Internal server error"
	if config.GetBool("app.debug") {
		message = truncateMessage(err.Error(), maxCloseMessageLength)
	}
	return c.internalClose(ws.CloseInternalServerErr, message)
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
		m := ws.FormatCloseMessage(code, message)
		writeErr := c.WriteControl(ws.CloseMessage, m, time.Now().Add(c.timeout))
		if writeErr != nil && !errors.Is(writeErr, ws.ErrCloseSent) {
			if strings.Contains(writeErr.Error(), "use of closed network connection") {
				err = writeErr
			}
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		defer cancel()

		select {
		case <-ctx.Done():
		case <-c.waitClose:
			close(c.waitClose)
		}

		err = c.Conn.Close()
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
