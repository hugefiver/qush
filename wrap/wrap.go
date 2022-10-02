package wrap

import (
	"crypto/tls"
	"net"

	"github.com/hugefiver/quic"
)

type QuicStream = quic.Stream

type Conn interface {
	net.Conn

	ConnectionStatus() ConnectionState
}

type QuicConn interface {
	Conn

	Stream() quic.Stream
}

type ConnectionState struct {
	tls.ConnectionState
	Used0RTT          bool
	SupportsDatagrams bool
}

type ConnWrapper struct {
	QuicStream
	Conn quic.Connection
}

func FromQuic(stream quic.Stream, conn quic.Connection) QuicConn {
	return &ConnWrapper{
		QuicStream: stream,
		Conn:       conn,
	}
}

func (c ConnWrapper) LocalAddr() net.Addr {
	return c.Conn.LocalAddr()
}

func (c ConnWrapper) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c ConnWrapper) Stream() quic.Stream {
	return c.QuicStream
}

func (c ConnWrapper) ConnectionStatus() ConnectionState {
	s := c.Conn.ConnectionState()

	return ConnectionState{
		ConnectionState:   s.TLS.ConnectionState,
		Used0RTT:          s.TLS.Used0RTT,
		SupportsDatagrams: s.SupportsDatagrams,
	}
}
