package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"golang.org/x/crypto/ssh"

	"encoding/binary"

	gopty "github.com/creack/pty"
)

func handleSSHChannel(c ssh.NewChannel, user string, shell string) {
	channel, reqs, err := c.Accept()
	if err != nil {
		log.Println("cannot accept channel:", err)
		return
	}

	pty, tty, err := gopty.Open()

	if err != nil {
		log.Println("cannot open pty:", err)
	}

	running := false

	for req := range reqs {
		log.Printf("Type: %v; Payload: %v", req.Type, req.Payload)
		ok := false

		home := os.Getenv("HOME")
		path := os.Getenv("PATH")

		envs := []string{
			"TERM=xterm",
			fmt.Sprintf("HOME=%s", home),
			fmt.Sprintf("PWD=%s", home),
			fmt.Sprintf("PATH=%s", path),
			fmt.Sprintf("USER=%s", user),
		}

		switch req.Type {
		case "exec":
			if running {
				req.Reply(false, nil)
				log.Println("ignore duplicate execute request")
				continue
			}
			running = true
			ok = true

			length := req.Payload[3]
			command := string(req.Payload[4 : length+4])
			cmd := exec.Command(shell, []string{"-c", command}...)
			cmd.Dir = home
			cmd.Env = envs

			err := PtyRun(cmd, tty)
			if err != nil {
				log.Printf("could not start command (%s)", err)
				continue
			}

			// set pipe of ssh channel and pty
			PipeChannels(channel, pty)

			// teardown session
			go func() {
				_, err := cmd.Process.Wait()
				if err != nil {
					log.Printf("failed to exit command: (%s)", err)
				}
				channel.Close()
				log.Printf("session closed")
			}()
		case "shell":
			if running {
				req.Reply(false, nil)
				log.Println("ignore duplicate execute request")
				continue
			}
			running = true
			ok = true

			cmd := exec.Command(shell)
			cmd.Dir = home
			cmd.Env = []string{
				"TERM=xterm",
				fmt.Sprintf("HOME=%s", home),
				fmt.Sprintf("PWD=%s", home),
				fmt.Sprintf("PATH=%s", path),
			}
			err := PtyRun(cmd, tty)
			if err != nil {
				log.Printf("%s", err)
			}

			PipeChannels(channel, pty)

			go func() {
				_, err := cmd.Process.Wait()
				if err != nil {
					log.Printf("failed to exit shell (%s)", err)
				}
				channel.Close()
				log.Printf("session closed")
			}()

			ok = true

		case "pty-req":
			// Responding 'ok' here will let the client
			// know we have a pty ready for input
			ok = true
			// Parse body...
			termLen := req.Payload[3]
			termEnv := string(req.Payload[4 : termLen+4])
			w, h := parseDims(req.Payload[termLen+4:])
			SetWinSize(pty.Fd(), w, h)
			log.Printf("pty-req '%s'\n", termEnv)
		case "window-change":
			w, h := parseDims(req.Payload)
			SetWinSize(pty.Fd(), w, h)
			continue //no response
		}

		if !ok {
			log.Printf("declining %s request\n", req.Type)
		}

		if req.WantReply {
			_ = req.Reply(ok, nil)
		}

	}

}

// parseDims extracts two uint32s from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

// WinSize stores the Height and Width of a terminal.
type WinSize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

func logAuthLog(conn ssh.ConnMetadata, method string, err error) {
	switch method {
	case "password":
		if err != nil {
			log.Printf("Failed to auth user %s login from %v using %s: %v\n",
				conn.User(), conn.RemoteAddr(), method, err)
		} else {
			log.Printf("Succeed to auth user %s login from %v using %s\n",
				conn.User(), conn.RemoteAddr(), method)
		}
	default:
		return
	}
}
