package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"

	gopty "github.com/creack/pty"
	"github.com/hugefiver/qush/ssh"
	"github.com/rs/zerolog/log"
)

func handleSSHChannel(c ssh.NewChannel, user string) {
	channel, reqs, err := c.Accept()
	if err != nil {
		log.Debug().Err(err).Msg("cannot accept channel")
		return
	}

	pty, tty, err := gopty.Open()

	if err != nil {
		log.Debug().Err(err).Msg("cannot open pty")
	}

	shell := programConfig.Shell
	if shell == "" {
		shell = os.Getenv("SHELL")
		log.Debug().Msgf("show env: SHELL=%s", shell)
		if shell == "" {
			shell = "sh"
			log.Debug().Msgf("SHELL is empty, will use `%s` ", shell)
		}
	}

	running := false

	for req := range reqs {
		log.Debug().Msgf("Type: %v; Payload: %v", req.Type, req.Payload)
		ok := false

		home := os.Getenv("HOME")
		path := os.Getenv("PATH")

		envs := []string{
			"TERM=xterm",
			fmt.Sprintf("HOME=%s", home),
			fmt.Sprintf("PWD=%s", home),
			fmt.Sprintf("PATH=%s", path),
			fmt.Sprintf("USER=%s", user),
			fmt.Sprintf("SHELL=%s", shell),
		}

		switch req.Type {
		case "exec":
			if running {
				req.Reply(false, nil)
				log.Debug().Msg("ignore duplicate execute request")
				continue
			}
			running = true
			ok = true

			length := req.Payload[3]
			command := string(req.Payload[4 : length+4])
			cmd := exec.Command(shell, []string{"-c", command}...)
			cmd.Dir = home
			cmd.Env = envs

			//e := func(e, v string) string {
			//	return fmt.Sprintf("%s=%s", strings.ToUpper(e), v)
			//}
			//envs := []string{
			//	e("USER", os.Getenv("USER")),
			//	e("HOME", os.Getenv("HOME")),
			//	e("PATH", os.Getenv("PATH")),
			//	e("PWD", os.Getenv("HOME")),
			//}
			//
			//cmd.Env = envs

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
				log.Debug().Msg("ignore duplicate execute request")
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
			log.Printf("pty-req '%s'", termEnv)
		case "window-change":
			w, h := parseDims(req.Payload)
			SetWinSize(pty.Fd(), w, h)
			continue //no response
		}

		if !ok {
			log.Printf("declining %s request", req.Type)
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
			log.Info().Err(err).
				Msgf("Failed to auth user %s login from %v using %s",
					conn.User(), conn.RemoteAddr(), method)
		} else {
			log.Info().
				Msgf("Succeed to auth user %s login from %v using %s",
					conn.User(), conn.RemoteAddr(), method)
		}
	default:
		return
	}
}
