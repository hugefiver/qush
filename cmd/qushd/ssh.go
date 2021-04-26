package main

import (
	"encoding/binary"
	"io"
	"os"
	"os/exec"
	"sync"

	gopty "github.com/creack/pty"
	"github.com/hugefiver/qush/ssh"
	"github.com/rs/zerolog/log"
)

func handleSSHChannel(c ssh.NewChannel) {
	channel, reqs, err := c.Accept()
	if err != nil {
		log.Debug().Err(err).Msg("cannot accept channel")
		return
	}

	pty, tty, err := gopty.Open()
	if err != nil {
		log.Debug().Err(err).Msg("cannot open pty")
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	for req := range reqs {
		log.Debug().Msgf("Type: %v; Payload: %v", req.Type, req.Payload)
		ok := false

		switch req.Type {
		case "exec":
			ok = true
			command := string(req.Payload[4 : req.Payload[3]+4])
			cmd := exec.Command(shell, []string{"-c", command}...)

			cmd.Stdout = channel
			cmd.Stderr = channel
			cmd.Stdin = channel

			err := cmd.Start()
			if err != nil {
				log.Printf("could not start command (%s)", err)
				continue
			}

			// teardown session
			go func() {
				_, err := cmd.Process.Wait()
				if err != nil {
					log.Printf("failed to exit bash (%s)", err)
				}
				channel.Close()
				log.Printf("session closed")
			}()
		case "shell":
			cmd := exec.Command(shell)
			cmd.Env = []string{"TERM=xterm"}
			err := PtyRun(cmd, tty)
			if err != nil {
				log.Printf("%s", err)
			}

			// Teardown session
			var once sync.Once
			closeFun := func() {
				channel.Close()
				log.Printf("session closed")
			}

			// Pipe session to bash and visa-versa
			go func() {
				io.Copy(channel, pty)
				once.Do(closeFun)
			}()

			go func() {
				io.Copy(pty, channel)
				once.Do(closeFun)
			}()

			// We don't accept any commands (Payload),
			// only the default shell.
			if len(req.Payload) == 0 {
				ok = true
			}
		case "pty-req":
			// Responding 'ok' here will let the client
			// know we have a pty ready for input
			ok = true
			// Parse body...
			termLen := req.Payload[3]
			termEnv := string(req.Payload[4 : termLen+4])
			w, h := parseDims(req.Payload[termLen+4:])
			SetWinSize(pty.Fd(), w, h)
			log.Printf("pty-req '%s'", termEnv)
		case "window-change":
			w, h := parseDims(req.Payload)
			SetWinSize(pty.Fd(), w, h)
			continue //no response
		}

		if !ok {
			log.Printf("declining %s request...", req.Type)
		}

		_ = req.Reply(ok, nil)
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
	if err != nil {
		log.Info().Err(err).Msgf("Failed to auth user %s login from %v using %s", conn.User(), conn.RemoteAddr(), method)
	} else {
		log.Info().Msgf("Succeed to auth login from %v using %s", conn.RemoteAddr(), method)
	}
}
