package ssh

import (
	"errors"
	"net"

	"github.com/hugefiver/qush/wrap"
)

// A connection represents an incoming connection.
type rawConnetion struct {
	transport *handshakeTransport
	rawConn

	// The connection protocol.
	*mux
}

// handshake performs key exchange and user authentication.
func (s *rawConnetion) serverHandshake(config *ServerConfig) (*Permissions, error) {
	if len(config.hostKeys) == 0 {
		return nil, errors.New("ssh: server has no host keys")
	}

	if !config.NoClientAuth && config.PasswordCallback == nil && config.PublicKeyCallback == nil &&
		config.KeyboardInteractiveCallback == nil && (config.GSSAPIWithMICConfig == nil ||
		config.GSSAPIWithMICConfig.AllowLogin == nil || config.GSSAPIWithMICConfig.Server == nil) {
		return nil, errors.New("ssh: no authentication methods configured but NoClientAuth is also false")
	}

	if config.ServerVersion != "" {
		s.serverVersion = []byte(config.ServerVersion)
	} else {
		s.serverVersion = []byte(packageVersion)
	}
	var err error
	s.clientVersion, err = exchangeVersions(s.rawConn.conn, s.serverVersion)
	if err != nil {
		return nil, err
	}

	tr := newTransport(s.rawConn.conn, config.Rand, false /* not client */)
	s.transport = newServerTransport(tr, s.clientVersion, s.serverVersion, config)

	if err := s.transport.waitSession(); err != nil {
		return nil, err
	}

	// We just did the key change, so the session ID is established.
	s.sessionID = s.transport.getSessionID()

	var packet []byte
	if packet, err = s.transport.readPacket(); err != nil {
		return nil, err
	}

	var serviceRequest serviceRequestMsg
	if err = Unmarshal(packet, &serviceRequest); err != nil {
		return nil, err
	}
	if serviceRequest.Service != serviceUserAuth {
		return nil, errors.New("ssh: requested service '" + serviceRequest.Service + "' before authenticating")
	}
	serviceAccept := serviceAcceptMsg{
		Service: serviceUserAuth,
	}
	if err := s.transport.writePacket(Marshal(&serviceAccept)); err != nil {
		return nil, err
	}

	perms, err := s.serverAuthenticate(config)
	if err != nil {
		return nil, err
	}
	s.mux = newMux(s.transport)
	return perms, err
}

func (s *rawConnetion) auth(config *ServerConfig) (*Permissions, error) {
	return &Permissions{
		CriticalOptions: nil,
		Extensions:      nil,
	}, nil
}

func (s *rawConnetion) serverAuthenticate(config *ServerConfig) (*Permissions, error) {
	return nil, nil
}

//func (s *rawConnetion) serverAuthenticate(config *ServerConfig) (*Permissions, error) {
//	sessionID := s.transport.getSessionID()
//	var cache pubKeyCache
//	var perms *Permissions
//
//	authFailures := 0
//	var authErrs []error
//	var displayedBanner bool
//
//userAuthLoop:
//	for {
//		if authFailures >= config.MaxAuthTries && config.MaxAuthTries > 0 {
//			discMsg := &disconnectMsg{
//				Reason:  2,
//				Message: "too many authentication failures",
//			}
//
//			if err := s.transport.writePacket(Marshal(discMsg)); err != nil {
//				return nil, err
//			}
//
//			return nil, discMsg
//		}
//
//		var userAuthReq userAuthRequestMsg
//		if packet, err := s.transport.readPacket(); err != nil {
//			if err == io.EOF {
//				return nil, &ServerAuthError{Errors: authErrs}
//			}
//			return nil, err
//		} else if err = Unmarshal(packet, &userAuthReq); err != nil {
//			return nil, err
//		}
//
//		if userAuthReq.Service != serviceSSH {
//			return nil, errors.New("ssh: client attempted to negotiate for unknown service: " + userAuthReq.Service)
//		}
//
//		s.user = userAuthReq.User
//
//		if !displayedBanner && config.BannerCallback != nil {
//			displayedBanner = true
//			msg := config.BannerCallback(s)
//			if msg != "" {
//				bannerMsg := &userAuthBannerMsg{
//					Message: msg,
//				}
//				if err := s.transport.writePacket(Marshal(bannerMsg)); err != nil {
//					return nil, err
//				}
//			}
//		}
//
//		perms = nil
//		authErr := ErrNoAuth
//
//		switch userAuthReq.Method {
//		case "none":
//			if config.NoClientAuth {
//				authErr = nil
//			}
//
//			// allow initial attempt of 'none' without penalty
//			if authFailures == 0 {
//				authFailures--
//			}
//		case "password":
//			if config.PasswordCallback == nil {
//				authErr = errors.New("ssh: password auth not configured")
//				break
//			}
//			payload := userAuthReq.Payload
//			if len(payload) < 1 || payload[0] != 0 {
//				return nil, parseError(msgUserAuthRequest)
//			}
//			payload = payload[1:]
//			password, payload, ok := parseString(payload)
//			if !ok || len(payload) > 0 {
//				return nil, parseError(msgUserAuthRequest)
//			}
//
//			perms, authErr = config.PasswordCallback(s, password)
//		case "keyboard-interactive":
//			if config.KeyboardInteractiveCallback == nil {
//				authErr = errors.New("ssh: keyboard-interactive auth not configured")
//				break
//			}
//
//			prompter := &sshClientKeyboardInteractive{s}
//			perms, authErr = config.KeyboardInteractiveCallback(s, prompter.Challenge)
//		case "publickey":
//			if config.PublicKeyCallback == nil {
//				authErr = errors.New("ssh: publickey auth not configured")
//				break
//			}
//			payload := userAuthReq.Payload
//			if len(payload) < 1 {
//				return nil, parseError(msgUserAuthRequest)
//			}
//			isQuery := payload[0] == 0
//			payload = payload[1:]
//			algoBytes, payload, ok := parseString(payload)
//			if !ok {
//				return nil, parseError(msgUserAuthRequest)
//			}
//			algo := string(algoBytes)
//			if !isAcceptableAlgo(algo) {
//				authErr = fmt.Errorf("ssh: algorithm %q not accepted", algo)
//				break
//			}
//
//			pubKeyData, payload, ok := parseString(payload)
//			if !ok {
//				return nil, parseError(msgUserAuthRequest)
//			}
//
//			pubKey, err := ParsePublicKey(pubKeyData)
//			if err != nil {
//				return nil, err
//			}
//
//			candidate, ok := cache.get(s.user, pubKeyData)
//			if !ok {
//				candidate.user = s.user
//				candidate.pubKeyData = pubKeyData
//				candidate.perms, candidate.result = config.PublicKeyCallback(s, pubKey)
//				if candidate.result == nil && candidate.perms != nil && candidate.perms.CriticalOptions != nil && candidate.perms.CriticalOptions[sourceAddressCriticalOption] != "" {
//					candidate.result = checkSourceAddress(
//						s.RemoteAddr(),
//						candidate.perms.CriticalOptions[sourceAddressCriticalOption])
//				}
//				cache.add(candidate)
//			}
//
//			if isQuery {
//				// The client can query if the given public key
//				// would be okay.
//
//				if len(payload) > 0 {
//					return nil, parseError(msgUserAuthRequest)
//				}
//
//				if candidate.result == nil {
//					okMsg := userAuthPubKeyOkMsg{
//						Algo:   algo,
//						PubKey: pubKeyData,
//					}
//					if err = s.transport.writePacket(Marshal(&okMsg)); err != nil {
//						return nil, err
//					}
//					continue userAuthLoop
//				}
//				authErr = candidate.result
//			} else {
//				sig, payload, ok := parseSignature(payload)
//				if !ok || len(payload) > 0 {
//					return nil, parseError(msgUserAuthRequest)
//				}
//				// Ensure the public key algo and signature algo
//				// are supported.  Compare the private key
//				// algorithm name that corresponds to algo with
//				// sig.Format.  This is usually the same, but
//				// for certs, the names differ.
//				if !isAcceptableAlgo(sig.Format) {
//					authErr = fmt.Errorf("ssh: algorithm %q not accepted", sig.Format)
//					break
//				}
//				signedData := buildDataSignedForAuth(sessionID, userAuthReq, algoBytes, pubKeyData)
//
//				if err := pubKey.Verify(signedData, sig); err != nil {
//					return nil, err
//				}
//
//				authErr = candidate.result
//				perms = candidate.perms
//			}
//		case "gssapi-with-mic":
//			if config.GSSAPIWithMICConfig == nil {
//				authErr = errors.New("ssh: gssapi-with-mic auth not configured")
//				break
//			}
//			gssapiConfig := config.GSSAPIWithMICConfig
//			userAuthRequestGSSAPI, err := parseGSSAPIPayload(userAuthReq.Payload)
//			if err != nil {
//				return nil, parseError(msgUserAuthRequest)
//			}
//			// OpenSSH supports Kerberos V5 mechanism only for GSS-API authentication.
//			if userAuthRequestGSSAPI.N == 0 {
//				authErr = fmt.Errorf("ssh: Mechanism negotiation is not supported")
//				break
//			}
//			var i uint32
//			present := false
//			for i = 0; i < userAuthRequestGSSAPI.N; i++ {
//				if userAuthRequestGSSAPI.OIDS[i].Equal(krb5Mesh) {
//					present = true
//					break
//				}
//			}
//			if !present {
//				authErr = fmt.Errorf("ssh: GSSAPI authentication must use the Kerberos V5 mechanism")
//				break
//			}
//			// Initial server response, see RFC 4462 section 3.3.
//			if err := s.transport.writePacket(Marshal(&userAuthGSSAPIResponse{
//				SupportMech: krb5OID,
//			})); err != nil {
//				return nil, err
//			}
//			// Exchange token, see RFC 4462 section 3.4.
//			packet, err := s.transport.readPacket()
//			if err != nil {
//				return nil, err
//			}
//			userAuthGSSAPITokenReq := &userAuthGSSAPIToken{}
//			if err := Unmarshal(packet, userAuthGSSAPITokenReq); err != nil {
//				return nil, err
//			}
//			authErr, perms, err = gssExchangeToken(gssapiConfig, userAuthGSSAPITokenReq.Token, s, sessionID,
//				userAuthReq)
//			if err != nil {
//				return nil, err
//			}
//		default:
//			authErr = fmt.Errorf("ssh: unknown method %q", userAuthReq.Method)
//		}
//
//		authErrs = append(authErrs, authErr)
//
//		if config.AuthLogCallback != nil {
//			config.AuthLogCallback(s, userAuthReq.Method, authErr)
//		}
//
//		if authErr == nil {
//			break userAuthLoop
//		}
//
//		authFailures++
//
//		var failureMsg userAuthFailureMsg
//		if config.PasswordCallback != nil {
//			failureMsg.Methods = append(failureMsg.Methods, "password")
//		}
//		if config.PublicKeyCallback != nil {
//			failureMsg.Methods = append(failureMsg.Methods, "publickey")
//		}
//		if config.KeyboardInteractiveCallback != nil {
//			failureMsg.Methods = append(failureMsg.Methods, "keyboard-interactive")
//		}
//		if config.GSSAPIWithMICConfig != nil && config.GSSAPIWithMICConfig.Server != nil &&
//			config.GSSAPIWithMICConfig.AllowLogin != nil {
//			failureMsg.Methods = append(failureMsg.Methods, "gssapi-with-mic")
//		}
//
//		if len(failureMsg.Methods) == 0 {
//			return nil, errors.New("ssh: no authentication methods configured but NoClientAuth is also false")
//		}
//
//		if err := s.transport.writePacket(Marshal(&failureMsg)); err != nil {
//			return nil, err
//		}
//	}
//
//	if err := s.transport.writePacket([]byte{msgUserAuthSuccess}); err != nil {
//		return nil, err
//	}
//	return perms, nil
//}

func (s *rawConnetion) Close() error {
	return s.rawConn.conn.Close()
}

// sshconn provides net.Conn metadata, but disallows direct reads and
// writes.
type rawConn struct {
	conn *wrap.ConnWrapper

	user          string
	sessionID     []byte
	clientVersion []byte
	serverVersion []byte
}

func exdup(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func (c *rawConn) User() string {
	return c.user
}

func (c *rawConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *rawConn) Close() error {
	return c.conn.Close()
}

func (c *rawConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *rawConn) SessionID() []byte {
	return exdup(c.sessionID)
}

func (c *rawConn) ClientVersion() []byte {
	return exdup(c.clientVersion)
}

func (c *rawConn) ServerVersion() []byte {
	return exdup(c.serverVersion)
}
