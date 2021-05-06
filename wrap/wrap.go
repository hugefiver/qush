package wrap

import (
	"crypto/tls"
	"net"

	"github.com/hugefiver/qush/quic"
)

type QuicStream = quic.Stream

type Conn interface {
	net.Conn

	Session() quic.Session
	Stream() quic.Stream
	ConnectionStatus() ConnectionState
}

type ConnectionState struct {
	tls.ConnectionState
	Used0RTT          bool
	SupportsDatagrams bool
}

type ConnWrapper struct {
	QuicStream
	session quic.Session
}

func From(stream quic.Stream, session quic.Session) Conn {
	return &ConnWrapper{
		QuicStream: stream,
		session:    session,
	}
}

func (c ConnWrapper) LocalAddr() net.Addr {
	return c.session.LocalAddr()
}

func (c ConnWrapper) RemoteAddr() net.Addr {
	return c.session.RemoteAddr()
}

func (c ConnWrapper) Session() quic.Session {
	return c.session
}

func (c ConnWrapper) Stream() quic.Stream {
	return c.QuicStream
}

func (c ConnWrapper) ConnectionStatus() ConnectionState {
	s := c.session.ConnectionState()
	t := s.TLS

	return ConnectionState{
		ConnectionState:   t.ConnectionState,
		Used0RTT:          t.Used0RTT,
		SupportsDatagrams: s.SupportsDatagrams,
	}
}
