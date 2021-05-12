package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/mattn/go-colorable"
	"golang.org/x/term"

	"golang.org/x/crypto/ssh"
)

func main() {
	user := flag.String("l", "", "login user")
	passwd := flag.String("passwd", "", "password of user")
	port := flag.Int("p", 22, "remote port")
	flag.Parse()

	if flag.NArg() == 0 {
		panic("need remote host")
	}
	args := flag.Args()
	host := args[0]
	commands := args[1:]

	config := &ssh.ClientConfig{
		Config:            ssh.Config{},
		User:              *user,
		Auth:              []ssh.AuthMethod{ssh.Password(*passwd)},
		HostKeyCallback:   ssh.InsecureIgnoreHostKey(),
		BannerCallback:    ssh.BannerDisplayStderr(),
		ClientVersion:     "SSH-2.0-TEST",
		HostKeyAlgorithms: nil,
		Timeout:           0,
	}

	// connect to server
	remote := fmt.Sprintf("%s:%d", host, *port)
	conn, err := net.Dial("tcp", remote)
	if err != nil {
		panic(err)
	}
	log.Println("connected to", remote)

	// ssh session
	c, channels, reqs, err := ssh.NewClientConn(conn, remote, config)
	if err != nil {
		panic(err)
	}
	log.Println("ssh connection created")

	client := ssh.NewClient(c, channels, reqs)
	sess, err := client.NewSession()
	if err != nil {
		panic(err)
	}

	infd := int(os.Stdin.Fd())
	if term.IsTerminal(infd) {
		oldStdinPerm, err := term.MakeRaw(infd)
		if err != nil {
			panic(err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldStdinPerm)
	}

	sess.Stdin = os.Stdin
	if term.IsTerminal(int(os.Stdout.Fd())) {
		ct := colorable.NewColorable(os.Stdout)
		sess.Stdout = ct
		sess.Stderr = ct
	} else {
		sess.Stdout = os.Stdout
		sess.Stderr = os.Stderr
	}

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}

	if err := sess.RequestPty("xterm", h, w, ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		panic(err)
	}

	if len(commands) == 0 {
		if err := sess.Shell(); err != nil {
			panic(err)
		}
		sess.Wait()
	} else {
		sess.Run(strings.Join(commands, " "))
	}
	log.Println("connect finished")
}
