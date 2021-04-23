package wrap

import (
	"net"

	"github.com/hugefiver/qush/quic"
)

type ConnWrapper struct {
	quic.Stream
	session quic.Session
}

func From(stream quic.Stream, session quic.Session) *ConnWrapper {
	return &ConnWrapper{
		Stream:  stream,
		session: session,
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
