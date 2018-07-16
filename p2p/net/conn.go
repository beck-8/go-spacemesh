package net

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/spacemeshos/go-spacemesh/crypto"
	"github.com/spacemeshos/go-spacemesh/p2p/delimited"
	"github.com/spacemeshos/go-spacemesh/p2p/net/wire"
	"gopkg.in/op/go-logging.v1"
)

var (
	ErrClosedChannel    = errors.New("unexpected closed connection channel")
	ErrConnectionClosed = errors.New("connections was intentionally closed")
)

// ConnectionSource specifies the connection originator - local or remote node.
type ConnectionSource int

// ConnectionSource values
const (
	Local ConnectionSource = iota
	Remote
)

// A network connection supporting full-duplex messaging
// Connection is an io.Writer and an io.Closer
type Connection struct {
	logger *logging.Logger
	// metadata for logging / debugging
	id        string           // uuid for logging
	source    ConnectionSource // remote or local
	created   time.Time
	remotePub crypto.PublicKey

	closeChan chan struct{}

	formatter wire.Formatter // format messages in some way

	incmoingMessages      []chan wire.InMessage
	incmoingMessagesMutex sync.RWMutex

	conn net.Conn // wrapped network connection
	net  Net      // network context

	session NetworkSession
}

// Create a new connection wrapping a net.Conn with a provided connection manager
func newConnection(conn net.Conn, n Net, s ConnectionSource, remotePub crypto.PublicKey) *Connection {

	// todo pass wire format inside and make it pluggable
	// todo parametrize channel size - hard-coded for now
	connection := &Connection{
		logger:    n.GetLogger(),
		id:        crypto.UUIDString(),
		created:   time.Now(),
		remotePub: remotePub,
		formatter: delimited.NewChan(10),
		source:    s,
		conn:      conn,
		net:       n,
		closeChan: make(chan struct{}),
	}

	connection.formatter.Pipe(connection.conn)

	// start processing channel-based message
	//TODO should be called explicitly by net
	//go connection.beginEventProcessing()
	return connection
}

func (c *Connection) ID() string {
	return c.id
}

// RemoteAddr returns the remote network address.
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c Connection) RemotePublicKey() crypto.PublicKey {
	return c.remotePub
}

func (c Connection) Source() ConnectionSource {
	return c.source
}

func (c *Connection) Session() NetworkSession {
	return c.session
}

func (c *Connection) String() string {
	return c.id
}

func (c *Connection) Subscribe() chan wire.InMessage {
	imc := make(chan wire.InMessage, 20)
	c.incmoingMessagesMutex.Lock()
	c.incmoingMessages = append(c.incmoingMessages, imc)
	c.incmoingMessagesMutex.Unlock()
	return imc
}

func (c *Connection) publish(im wire.InMessage) {
	c.incmoingMessagesMutex.RLock()
	for _, imc := range c.incmoingMessages {
		imc <- im
	}
	c.incmoingMessagesMutex.RUnlock()
}

// Send binary data to a connection
// data is copied over so caller can get rid of the data
// Concurrency: can be called from any go routine
func (c *Connection) Send(m []byte) error {
	return wire.Send(c.formatter, m)
}

// Close closes the connection (implements io.Closer). It is go safe.
func (c *Connection) Close() {
	c.closeChan <- struct{}{}
}

// Push outgoing message to the connections
// Read from the incoming new messages and send down the connection
func (c *Connection) beginEventProcessing() {

	var err error

Loop:
	for {
		select {
		case msg, ok := <-c.formatter.In():

			if !ok { // chan closed
				err = ErrClosedChannel
				break Loop
			}

			if msg.Error() != nil {
				err = msg.Error()
				break Loop
			}

			if c.session == nil {
				c.net.GetLogger().Info("DEBUG: got pre session message")
				err = c.net.HandlePreSessionIncomingMessage(c, msg)
				if err != nil {
					break Loop
				}
			} else {
				// channel for protocol messages
				go c.publish(msg)
				//c.net.GetIncomingMessage() <- IncomingMessage{c, msg.Message()}
			}

		case <-c.closeChan:
			err = ErrConnectionClosed
			break Loop
		}
	}

	c.formatter.Close()
	c.publish(c.formatter.MakeIn(nil, err))
	for _, cim := range c.incmoingMessages {
		close(cim)
	}
	// TODO: Teardown this connection
}
