package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	golog "log"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/onsi/ginkgo/reporters/stenographer/support/go-colorable"

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
	quicConfig := &quic.Config{
		Versions:                       nil,
		ConnectionIDLength:             0,
		HandshakeIdleTimeout:           0,
		MaxIdleTimeout:                 0,
		AcceptToken:                    nil,
		TokenStore:                     nil,
		InitialStreamReceiveWindow:     0,
		MaxStreamReceiveWindow:         0,
		InitialConnectionReceiveWindow: 0,
		MaxConnectionReceiveWindow:     0,
		MaxIncomingStreams:             0,
		MaxIncomingUniStreams:          0,
		StatelessResetKey:              nil,
		KeepAlive:                      true,
		DisablePathMTUDiscovery:        false,
		EnableDatagrams:                false,
		Tracer:                         nil,
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill)

		for s := range ch {
			golog.Fatalf("get a %s signal, program will exit.", s.String())
		}
	}()

	// connect server
	session, err := quic.DialAddr("10.64.202.123:22", tlsConfig, quicConfig)
	if err != nil {
		golog.Fatal(err)
	}
	golog.Println("connected")

	// open a QUIC stream
	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		golog.Fatal(err)
	}
	golog.Println("opened a QUIC stream")

	user := "test"
	config := &ssh.ClientConfig{
		Config: ssh.Config{},
		User:   user,
		//Auth:              []ssh.AuthMethod{ssh.Password("test")},
		Auth:              []ssh.AuthMethod{EnterPasswd(user)},
		HostKeyCallback:   hostKeyConfirm,
		BannerCallback:    ssh.BannerDisplayStderr(),
		ClientVersion:     clientVersion,
		HostKeyAlgorithms: nil,
		Timeout:           0,
	}

	conn, channels, reqs, err := ssh.NewClientConn(wrap.From(stream, session), session.RemoteAddr().String(), config)
	if err != nil {
		golog.Fatal(err)
	}
	golog.Println("new SSH conn created")

	client := ssh.NewClient(conn, channels, reqs)
	sshSession, err := client.NewSession()
	if err != nil {
		golog.Fatalln(err)
	}
	defer func() {
		sshSession.Close()
	}()

	//oldStdinPerm, err := term.MakeRaw(int(os.Stdin.Fd()))
	//if err != nil {
	//	golog.Fatalln(err)
	//}
	//defer term.Restore(int(os.Stdin.Fd()), oldStdinPerm)
	//
	//oldStdoutPerm, err := term.MakeRaw(int(os.Stdout.Fd()))
	//if err != nil {
	//	golog.Fatalln(err)
	//}
	//defer term.Restore(int(os.Stdout.Fd()), oldStdoutPerm)

	//ch := <-channels
	//c, _, err := ch.Accept()
	//if err != nil {
	//	golog.Fatalln(err)
	//}
	//
	//term.NewTerminal(c, "> ")

	sshSession.Stdin = os.Stdin
	//sshSession.Stdout = os.Stdout
	//sshSession.Stderr = os.Stdout
	ct := colorable.NewNonColorable(os.Stdout)
	sshSession.Stdout = ct
	sshSession.Stderr = ct
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		golog.Printf("window size: %dx%d, err=%v \n", w, h, err)
	} else {
		golog.Printf("windows size: %dx%d \n", w, h)
	}
	if err := sshSession.RequestPty("xterm", h, w, ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		golog.Fatalln(err)
	}

	golog.Println("gonna to run `ls` command")
	err = sshSession.Run("ls -l")
	//err = sshSession.Shell()
	if err != nil {
		golog.Fatalln(err)
	}
	//sshSession.Wait()
}

func verifyConnection(status tls.ConnectionState) error {
	return nil
}

func hostKeyConfirm(hostname string, remote net.Addr, key ssh.PublicKey) error {
	fmt.Printf("Host key fingerprint from %s(%v) is:\n", hostname, remote)
	finger := ssh.FingerprintSHA256(key)
	fmt.Printf("%s -> %s \n", key.Type(), finger)
	var in string
	for {
		fmt.Print("do you still want to connect(yes/no): ")
		fmt.Scan(&in)
		fmt.Println()
		switch strings.ToLower(strings.TrimSpace(in)) {
		case "yes":
			return nil
		case "no":
			return errors.New("connection canceled by user")
		default:
			fmt.Println(`invalid input, please input "yes" or "no". \n`)
			continue
		}
	}
}
