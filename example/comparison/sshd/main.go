package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/hugefiver/qush/key"

	"golang.org/x/crypto/ssh"
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

func main() {
	keypath := flag.String("key", "./key", "server key")
	bind := flag.String("bind", "0.0.0.0:22", "bind address")
	shell := flag.String("shell", "/bin/sh", "shell")

	flag.Parse()

	sshConfig := &ssh.ServerConfig{
		Config:                      ssh.Config{},
		NoClientAuth:                false,
		MaxAuthTries:                3,
		PasswordCallback:            PasswordAuthFunc,
		PublicKeyCallback:           nil,
		KeyboardInteractiveCallback: nil,
		AuthLogCallback:             nil,
		ServerVersion:               "SSH-2.0-TEST",
		BannerCallback:              nil,
		GSSAPIWithMICConfig:         nil,
	}

	// add host key
	k, err := key.LoadHostKey(*keypath)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.NewSignerFromKey(k)
	if err != nil {
		panic(err)
	}
	sshConfig.AddHostKey(signer)

	// bind address
	listener, err := net.Listen("tcp", *bind)
	if err != nil {
		panic(err)
	}
	log.Println("listen at", *bind)
	defer listener.Close()

	// serve loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("cannot accept connection from %v: %v\n",
				conn.RemoteAddr(), err)
			continue
		}
		go handleSSH(conn, sshConfig, *shell)
	}
}

func handleSSH(conn net.Conn, config *ssh.ServerConfig, shell string) {
	c, channels, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		log.Println("failed to start a ssh connection:", err)
		return
	}

	go ssh.DiscardRequests(reqs)

	// Service the incoming Channel channel.
	for newChannel := range channels {
		if t := newChannel.ChannelType(); t != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}

		go handleSSHChannel(newChannel, c.User(), shell)

	}
}
