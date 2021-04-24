package main

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/rs/zerolog/log"
)

// Start assigns a pseudo-terminal tty os.File to c.Stdin, c.Stdout,
// and c.Stderr, calls c.Start, and returns the File of the tty's
// corresponding pty.
func PtyRun(c *exec.Cmd, tty *os.File) (err error) {
	defer tty.Close()
	c.Stdout = tty
	c.Stdin = tty
	c.Stderr = tty
	c.SysProcAttr = &syscall.SysProcAttr{}
	return c.Start()
}

func SetWinSize(fd uintptr, w, h uint32) {
	log.Print("Windows not supported window resize")
}
