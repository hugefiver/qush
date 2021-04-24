package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	golog "log"
	"net"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/hugefiver/qush/quic"
	"github.com/hugefiver/qush/ssh"
	"github.com/hugefiver/qush/wrap"
)

var clientVersion = "QUSH-0.0.1"

func main() {
	tlsConfig := &tls.Config{
		VerifyPeerCertificate:       nil,
		VerifyConnection:            verifyConnection,
		NextProtos:                  []string{"qush"},
		InsecureSkipVerify:          true,
		CipherSuites:                nil,
		PreferServerCipherSuites:    false,
		SessionTicketsDisabled:      false,
		CurvePreferences:            nil,
		DynamicRecordSizingDisabled: false,
		Renegotiation:               0,
		KeyLogWriter:                nil,
	}
	session, err := quic.DialAddr("10.64.202.123:22", tlsConfig, nil)
	if err != nil {
		golog.Fatal(err)
	}
	golog.Println("connected")

	stream, err := session.OpenStream()
	if err != nil {
		golog.Fatal(err)
	}

	config := &ssh.ClientConfig{
		Config:            ssh.Config{},
		User:              "test",
		Auth:              nil,
		HostKeyCallback:   hostKeyConfirm,
		BannerCallback:    nil,
		ClientVersion:     clientVersion,
		HostKeyAlgorithms: nil,
		Timeout:           0,
	}
	conn, channels, reqs, err := ssh.NewClientConn(wrap.From(stream, session), session.RemoteAddr().String(), config)
	if err != nil {
		golog.Fatal(err)
	}

	_ = ssh.NewClient(conn, channels, reqs)

	oldStdinPerm, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		golog.Fatalln(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldStdinPerm)

	oldStdoutPerm, err := term.MakeRaw(int(os.Stdout.Fd()))
	if err != nil {
		golog.Fatalln(err)
	}
	defer term.Restore(int(os.Stdout.Fd()), oldStdoutPerm)

	ch := <-channels
	c, _, err := ch.Accept()
	if err != nil {
		golog.Fatalln(err)
	}

	term.NewTerminal(c, "> ")

}

func verifyConnection(status tls.ConnectionState) error {
	return nil
}

func hostKeyConfirm(hostname string, remote net.Addr, key ssh.PublicKey) error {
	fmt.Printf("Host key finger print from %s(%v) is: ", hostname, remote)
	finger := ssh.FingerprintSHA256(key)
	fmt.Println("-> ", finger)
	var in string
	for {
		fmt.Scanf("do you still want to connect(yes/no): %s", in)
		fmt.Println(in)
		switch strings.ToLower(strings.TrimSpace(in)) {
		case "yes":
			return nil
		case "no":
			return errors.New("connection canceled by user")
		default:
			continue
		}
	}
}
