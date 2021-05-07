package auth

import (
	"fmt"

	"github.com/hugefiver/qush/ssh"
)

func PasswordAuthFunc(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	// just for test
	if conn.User() == "test" && string(password) == "test" {
		return nil, nil
	} else {
		//log.Info().Msgf("Failed login with %s from %v", conn.User(), conn.RemoteAddr())
		return nil, fmt.Errorf("login failed for %s", conn.User())
	}
}
