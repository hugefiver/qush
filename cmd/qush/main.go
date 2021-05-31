package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	golog "log"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/pflag"

	"github.com/mattn/go-colorable"
	"golang.org/x/term"

	"github.com/hugefiver/qush/consits"
	"github.com/hugefiver/qush/quic"
	"github.com/hugefiver/qush/ssh"
	"github.com/hugefiver/qush/util"
	"github.com/hugefiver/qush/wrap"
)

var version = "0.0.1"
var buildTime = "unknown"
var clientVersion = "QUSH-0.0.1"

var debug *golog.Logger

func main() {
	flags := ParseFlags()

	if flags.Help {
		pflag.Usage()
		os.Exit(0)
	}

	if flags.Version {
		showVerbose()
		os.Exit(0)
	}

	if flags.Verbose > 0 {
		debug = golog.New(os.Stderr, "[debug]", golog.Lmsgprefix)
	} else {
		debug = golog.New(ioutil.Discard, "", 0)
	}

	tlsConfig := &tls.Config{
		VerifyPeerCertificate:       nil,
		VerifyConnection:            verifyConnection,
		NextProtos:                  []string{"qush"},
		InsecureSkipVerify:          true,
		CipherSuites:                []uint16{tls.TLS_AES_128_GCM_SHA256},
		PreferServerCipherSuites:    false,
		SessionTicketsDisabled:      false,
		CurvePreferences:            nil,
		DynamicRecordSizingDisabled: false,
		Renegotiation:               0,
		KeyLogWriter:                nil,
		MinVersion:                  tls.VersionTLS13,
	}
	quicConfig := &quic.Config{
		Versions:                       nil,
		ConnectionIDLength:             0,
		HandshakeIdleTimeout:           0,
		MaxIdleTimeout:                 0,
		AcceptToken:                    nil,
		TokenStore:                     nil,
		InitialStreamReceiveWindow:     8 * util.MB,
		MaxStreamReceiveWindow:         16 * util.MB,
		InitialConnectionReceiveWindow: 8 * util.MB,
		MaxConnectionReceiveWindow:     32 * util.MB,
		MaxIncomingStreams:             0,
		MaxIncomingUniStreams:          0,
		StatelessResetKey:              nil,
		KeepAlive:                      true,
		//DisablePathMTUDiscovery:        true,
		//EnableDatagrams:                true,
		Tracer: nil,
	}

	// clear closers when exist
	var closers []func()
	clearer := func() {
		cs := closers
		for _, f := range cs {
			f()
		}
	}

	addCloser := func(fn func()) {
		closers = append([]func(){fn}, closers...)
	}

	defer clearer()

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill)

		for s := range ch {
			debug.Printf("get a %s signal, program will exit.", s.String())
			clearer()
			os.Exit(0)
		}
	}()

	// connect server
	debug.Println("starting connect")
	session, err := quic.DialAddrEarly(fmt.Sprintf("%s:%d", flags.Host, flags.Port), tlsConfig, quicConfig)
	if err != nil {
		debug.Fatal(err)
	}
	addCloser(func() {
		session.CloseWithError(consits.DISCONNECT, "program exited")
	})
	debug.Println("connected")

	// open a QUIC stream
	stream, err := session.OpenStream()
	if err != nil {
		debug.Fatal(err)
	}
	addCloser(func() {
		stream.Close()
	})
	debug.Println("opened a QUIC stream")

	keyConfirmCallback := hostKeyConfirm
	if flags.IgnorePubKey {
		keyConfirmCallback = passHostKeyConfirm
	}

	// login user
	user := flags.User

	// password auth
	passwdAuth := EnterPasswd(user, flags.PasswdTries)
	if flags.DebugOnlyPasswd != nil {
		passwdAuth = ssh.Password(*flags.DebugOnlyPasswd)
	}

	config := &ssh.ClientConfig{
		Config:            ssh.Config{},
		User:              user,
		Auth:              []ssh.AuthMethod{passwdAuth},
		HostKeyCallback:   keyConfirmCallback,
		BannerCallback:    ssh.BannerDisplayStderr(),
		ClientVersion:     clientVersion,
		HostKeyAlgorithms: nil,
		Timeout:           0,
	}

	conn, channels, reqs, err := ssh.NewClientConn(wrap.From(stream, session), session.RemoteAddr().String(), config)
	if err != nil {
		debug.Fatal(err)
	}
	debug.Println("new SSH conn created")

	client := ssh.NewClient(conn, channels, reqs)
	sshSession, err := client.NewSession()
	if err != nil {
		debug.Fatalln(err)
	}
	defer func() {
		sshSession.Close()
	}()

	infd := int(os.Stdin.Fd())
	if term.IsTerminal(infd) {
		oldStdinPerm, err := term.MakeRaw(infd)
		if err != nil {
			debug.Fatalln(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldStdinPerm)
	}

	//oldStdoutPerm, err := term.MakeRaw(int(os.Stdout.Fd()))
	//if err != nil {
	//	debug.Fatalln(err)
	//}
	//defer term.Restore(int(os.Stdout.Fd()), oldStdoutPerm)

	sshSession.Stdin = os.Stdin
	if term.IsTerminal(int(os.Stdout.Fd())) {
		ct := colorable.NewColorable(os.Stdout)
		sshSession.Stdout = ct
		sshSession.Stderr = ct
	} else {
		sshSession.Stdout = os.Stdout
		sshSession.Stderr = os.Stderr
	}

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		debug.Printf("window size: %dx%d, err=%v \n", w, h, err)
	} else {
		debug.Printf("windows size: %dx%d \n", w, h)
	}

	if err := sshSession.RequestPty("xterm", h, w, ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		debug.Fatalln(err)
	}

	if len(flags.Cmd) == 0 {
		debug.Println("request a shell")
		if err := sshSession.Shell(); err != nil {
			debug.Fatalln(err)
		}
		sshSession.Wait()
	} else {
		sshSession.Run(strings.Join(flags.Cmd, " "))
	}
	debug.Println("connect finished")
}

func verifyConnection(status tls.ConnectionState) error {
	return nil
}

func passHostKeyConfirm(hostname string, remote net.Addr, key ssh.PublicKey) error {
	fmt.Printf("Host key fingerprint from %s(%v) is:\n", hostname, remote)
	finger := ssh.FingerprintSHA256(key)
	fmt.Printf("%s -> %s \n", key.Type(), finger)
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
		switch strings.ToLower(strings.TrimSpace(in)) {
		case "yes":
			return nil
		case "no":
			return errors.New("connection canceled by user")
		default:
			fmt.Println(`invalid input, please input "yes" or "no".`)
			continue
		}
	}
}

func showVerbose() {
	fmt.Println("QUSH - Quick UDP Shell")
	fmt.Printf("Client version %s, build time %s \n", version, buildTime)
	fmt.Println("Author Hugefiver<i@iruri.moe> 2021")
}

func init() {
	golog.SetFlags(golog.Lmsgprefix)
}
