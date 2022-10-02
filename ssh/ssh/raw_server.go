package ssh

import (
	"fmt"

	"github.com/hugefiver/qush/wrap"
)

func RawServer(c *wrap.ConnWrapper,
	config *ServerConfig) (*ServerConn, <-chan NewChannel, <-chan *Request, error) {

	fullConf := *config
	fullConf.SetDefaults()
	if fullConf.MaxAuthTries == 0 {
		fullConf.MaxAuthTries = 6
	}

	// Check if the config contains any unsupported key exchanges
	for _, kex := range fullConf.KeyExchanges {
		if _, ok := serverForbiddenKexAlgos[kex]; ok {
			return nil, nil, nil, fmt.Errorf("ssh: unsupported key exchange %s for server", kex)
		}
	}

	s := &rawConnetion{
		rawConn: rawConn{conn: c},
	}
	perms, err := s.serverHandshake(&fullConf)
	if err != nil {
		_ = c.Close()
		return nil, nil, nil, err
	}
	return &ServerConn{s, perms}, s.mux.incomingChannels, s.mux.incomingRequests, nil
}
