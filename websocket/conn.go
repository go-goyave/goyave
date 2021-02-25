package websocket

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/System-Glitch/goyave/v3"
	"github.com/System-Glitch/goyave/v3/config"

	ws "github.com/gorilla/websocket"
)

// Conn represents a WebSocket connection.
type Conn struct {
	*ws.Conn
	waitClose     chan struct{}
	receivedClose bool
	timeout       time.Duration
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

func (c *Conn) closeHandler(code int, text string) error {
	// No need to lock receivedClose because there can be at most one
	// open reader on a connection.
	if c.receivedClose {
		return nil
	}
	c.receivedClose = true
	c.waitClose <- struct{}{}
	return nil
}

// CloseNormal performs the closing handshake as specified by
// RFC 6455 Section 1.4. Sends status code 1000 (normal closure) and
// message "Server closed connection".
//
// Don't use this inside websocket Handler. Only the HTTP handler
// that upgraded the connection should call this function.
func (c *Conn) CloseNormal() error {
	return c.Close(ws.CloseNormalClosure, NormalClosureMessage)
}

// CloseWithError performs the closing handshake as specified by
// RFC 6455 Section 1.4 because a server error occurred.
// Sends status code 1011 (internal server error) and
// message "Internal server error". If debug is enabled,
// the message is set to the given error's message.
//
// Don't use this inside websocket Handler. Only the HTTP handler
// that upgraded the connection should call this function.
func (c *Conn) CloseWithError(err error) error {
	message := "Internal server error"
	if config.GetBool("app.debug") {
		message = truncateMessage(err.Error(), maxCloseMessageLength)
	}
	return c.Close(ws.CloseInternalServerErr, message)
}

// Close performs the closing handshake as specified by RFC 6455 Section 1.4.
//
// Don't use this inside websocket Handler. Only the HTTP handler
// that upgraded the connection should call this function.
func (c *Conn) Close(code int, message string) error {
	m := ws.FormatCloseMessage(code, message)
	err := c.WriteControl(ws.CloseMessage, m, time.Now().Add(c.timeout))
	if err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") {
			c.Conn.Close()
			return fmt.Errorf("%w. Don't close the connection manually, this prevents close handshake", err)
		}
		if errors.Is(err, ws.ErrCloseSent) {
			c.Conn.Close()
			return fmt.Errorf("%w. A close frame has been sent before the Handler returned, preventing close handshake", err)
		}
		return c.Conn.Close()
	}

	if !c.receivedClose {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
		// In this branch, we know the client has NOT initiated the close handshake.
		// Read until error.
		go func() {
			for {
				if _, _, err := c.ReadMessage(); err != nil {
					return
				}
			}
		}()

		select {
		case <-ctx.Done():
			goyave.ErrLogger.Println(ErrCloseTimeout)
		case <-c.waitClose:
			close(c.waitClose)
		}
	}

	// TODO properly shutdown before goyave.Start returns?
	return c.Conn.Close()
}
