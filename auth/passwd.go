package auth

import (
	"fmt"

	"github.com/hugefiver/qush/config"
	"github.com/hugefiver/qush/ssh"
)

var user, passwd string

func PasswordAuthFunc(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	// just for test
	if conn.User() == user && string(password) == passwd {
		return nil, nil
	} else {
		//log.Info().Msgf("Failed login with %s from %v", conn.User(), conn.RemoteAddr())
		return nil, fmt.Errorf("login failed for %s", conn.User())
	}
}

func Init() {
	c, err := config.GetServerConfig()
	if err != nil {
		panic(err)
	}

	user, passwd = c.DefaultUser, c.DefaultPasswd

	if user == "" || passwd == "" {
		user, passwd = "test", "test"
	}
}
